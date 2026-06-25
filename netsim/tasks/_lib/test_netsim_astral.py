"""Offline tests for netsim_astral -- no VM, no live astrald.

Exercises the interrogators against synthetic AstralObjects, parse_cli's
stream handling, and the Go-CLI fallback command construction. Run with:

    python3 -m unittest -v        # from this directory
"""
import os
import sys
import unittest

sys.path.insert(0, os.path.dirname(os.path.realpath(__file__)))
import netsim_astral as na  # noqa: E402  (also bootstraps astral onto sys.path)
import astral  # noqa: E402


def O(type, value=None):
    return astral.obj(type, value)


class InterrogatorTests(unittest.TestCase):
    def test_contract(self):
        objs = [O("mod.user.contract",
                  {"Contract": {"Contract": {"Issuer": "02aa", "Subject": "03bb"}}})]
        self.assertEqual(na.contract(objs), ("02aa", "03bb"))
        self.assertEqual(na.contract([O("x", {})]), (None, None))

    def test_linked_sibling_and_identities(self):
        objs = [O("s", {"Identity": "03bb", "Linked": True}),
                O("s", {"Identity": "03cc", "Linked": False})]
        self.assertEqual(na.linked_sibling(objs), "03bb")
        self.assertEqual(na.swarm_identities(objs), {"03bb", "03cc"})
        self.assertIsNone(na.linked_sibling([O("s", {"Identity": "03cc", "Linked": False})]))

    def test_has_link_to(self):
        objs = [O("l", {"RemoteIdentity": "03bb", "Network": "tcp"})]
        self.assertTrue(na.has_link_to(objs, "03bb"))
        self.assertFalse(na.has_link_to(objs, "03cc"))

    def test_is_expelled_nested(self):
        objs = [O("mod.user.signed_expulsion", {"Expulsion": {"Subject": "03bb"}})]
        self.assertTrue(na.is_expelled(objs, "03bb"))
        self.assertFalse(na.is_expelled(objs, "03cc"))
        # an error_message naming the id must not count as an expulsion record
        self.assertFalse(na.is_expelled([O("error_message", "03bb not found")], "03bb"))

    def test_loaded_payload_and_errors(self):
        objs = [O("error_message", "boom"), O("string8", "hello")]
        self.assertEqual(na.loaded_payload(objs), "hello")
        self.assertEqual(na.error_messages(objs), ["boom"])
        self.assertIsNone(na.loaded_payload([O("error_message", "boom")]))

    def test_tor_links_and_endpoint(self):
        objs = [O("l", {"Network": "tor", "RemoteIdentity": "03bb",
                        "RemoteEndpoint": {"Object": "abc.onion:1791"}}),
                O("l", {"Network": "tcp", "RemoteIdentity": "03cc"})]
        self.assertEqual(na.tor_links(objs), [("03bb", "abc.onion:1791")])
        self.assertEqual(na.endpoint_addr("x.onion"), "x.onion")
        self.assertEqual(na.endpoint_addr({"Object": "y.onion"}), "y.onion")
        self.assertEqual(na.endpoint_addr(None), "")

    def test_resolve_onion(self):
        objs = [O("e", {"Endpoint": "10.0.0.1:1791"}),
                O("e", {"Endpoint": {"Object": "abc.onion:1791"}})]
        self.assertEqual(na.resolve_onion(objs), "abc.onion:1791")
        self.assertIsNone(na.resolve_onion([O("e", {"Endpoint": "10.0.0.1:1791"})]))


class ParseCliTests(unittest.TestCase):
    def test_drops_eos_keeps_error(self):
        raw = ('{"Type":"string8","Object":"hi"}\n'
               '{"Type":"error_message","Object":"nope"}\n'
               '\n'
               'not-json\n'
               '{"Type":"eos","Object":null}\n')
        objs = na.parse_cli(raw)
        self.assertEqual([o.type for o in objs], ["string8", "error_message"])
        self.assertEqual(na.loaded_payload(objs), "hi")
        self.assertEqual(na.error_messages(objs), ["nope"])

    def test_empty(self):
        self.assertEqual(na.parse_cli(""), [])
        self.assertEqual(na.parse_cli(None), [])


class ShellRoutingTests(unittest.TestCase):
    """Node with no client must build the exact Go astral-query command."""

    def setUp(self):
        self.calls = []
        self._orig = na.ssh

        def fake_ssh(vm, remote):
            self.calls.append((vm, remote))
            return '{"Type":"string8","Object":"hi"}\n{"Type":"eos","Object":null}\n'

        na.ssh = fake_ssh

    def tearDown(self):
        na.ssh = self._orig

    def test_untokened(self):
        node = na.Node("node1", None, "")
        objs = node.call("user.info")
        self.assertEqual(self.calls[-1], ("node1", "astral-query user.info -out json"))
        self.assertEqual(na.loaded_payload(objs), "hi")

    def test_tokened_with_args(self):
        na.Node("node1", None, "TKN").call("objects.load", {"id": "X", "repo": "local"})
        self.assertEqual(
            self.calls[-1][1],
            "export ASTRALD_APPHOST_TOKEN=TKN; "
            "astral-query objects.load -id 'X' -repo 'local' -out json")

    def test_peer_target(self):
        na.Node("node1", None, "TKN").call("objects.load", {"id": "X"}, target="node2")
        self.assertEqual(
            self.calls[-1][1],
            "export ASTRALD_APPHOST_TOKEN=TKN; "
            "astral-query node2:objects.load -id 'X' -out json")

    def test_shell_ops_pin_forces_cli(self):
        # even with a (truthy sentinel) client, a pinned op must go to the shell
        na.SHELL_OPS.add("user.info")
        try:
            node = na.Node("node1", object(), "")
            node.call("user.info")
            self.assertEqual(self.calls[-1][1], "astral-query user.info -out json")
        finally:
            na.SHELL_OPS.discard("user.info")


if __name__ == "__main__":
    unittest.main()
