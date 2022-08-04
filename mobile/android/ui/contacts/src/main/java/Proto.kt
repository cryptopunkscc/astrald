package cc.cryptopunks.astral.ui.contacts

import cc.cryptopunks.astral.client.enc.NetworkEncoder
import cc.cryptopunks.astral.client.ext.query
import cc.cryptopunks.astral.client.ext.decodeList

suspend fun NetworkEncoder.getContacts(): List<Contact> = query("contacts") {
    decodeList<Contact>().sortedBy(Contact::id)
}

class Contact(
    val id: String,
    val name: String,
)
