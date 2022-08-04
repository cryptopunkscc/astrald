package cc.cryptopunks.astral.ui.contacts

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.LocalContentColor
import androidx.compose.material.MaterialTheme
import androidx.compose.material.Scaffold
import androidx.compose.material.Text
import androidx.compose.material.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalInspectionMode
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import cc.cryptopunks.astral.theme.AstralTheme
import com.google.accompanist.swiperefresh.SwipeRefresh
import com.google.accompanist.swiperefresh.rememberSwipeRefreshState
import kotlin.random.Random

@Composable
@Preview
private fun MainPreview() = AstralTheme {
    MainView(rememberContactsPreviewModel())
}

@Composable
fun rememberContactsPreviewModel(): ContactsModel = remember {
    ContactsModel().apply {
        contacts.value = (0..10).map {
            Contact(
                name = "Name $it",
                id = Random.nextLong(
                    from = 100000000000000000,
                    until = 999999999999999999,
                ).toString()
            )
        }
    }
}

@Composable
internal fun MainView(
    model: ContactsModel = viewModel(),
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text(text = "Astral contacts")
                },
            )
        },
        content = {
            ContactsView(
                modifier = Modifier.padding(it),
                model = model
            )
        }
    )
}


@Composable
fun ContactsView(
    modifier: Modifier = Modifier,
    model: ContactsModel = viewModel(),
) {
    if (!LocalInspectionMode.current) LaunchedEffect(Unit) { model.loadContacts() }

    val contacts by model.contacts.collectAsState()
    val refreshState = rememberSwipeRefreshState(true)
    refreshState.isRefreshing = model.loading.collectAsState().value
    SwipeRefresh(
        state = refreshState,
        onRefresh = { model.loadContacts() }
    ) {
        LazyColumn(
            modifier = modifier.fillMaxSize()
        ) {
            items(contacts) { contact ->
                HorizontalDivider {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable(model.selectable) { model.selected.tryEmit(contact) }
                            .padding(16.dp)
                    ) {
                        Text(
                            text = contact.name,
                            style = MaterialTheme.typography.subtitle1,
                        )
                        Spacer(modifier = Modifier.width(8.dp))
                        Text(
                            text = contact.id.takeLast(8).chunked(4).joinToString("-"),
                            style = MaterialTheme.typography.caption,
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun HorizontalDivider(
    content: @Composable () -> Unit,
) = Column {
    content()
    HorizontalDivider()
}

@Composable
fun HorizontalDivider() = Spacer(
    Modifier
        .background(LocalContentColor.current.copy(0.25f))
        .fillMaxWidth()
        .height(0.5.dp)
)
