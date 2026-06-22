# configure-astral-agent

Installs the `astral-agent` skill into the Qwen Code operator on node1, so it can
drive astrald from the skill's playbooks + astral-docs instead of from procedures
spelled out in each prompt. The host clones the private `satforgedev/skills` via a
deploy key (`SATFORGE_SKILLS_DEPLOY_KEY`, a host path to the private key) and links
the skill into `~tester/.qwen/skills/astral-agent`. Part of `lab.story`.
