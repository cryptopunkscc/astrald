package cc.cryptopunks.astral.ui.log

import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel

@Composable
fun LogView(
    modifier: Modifier = Modifier,
    logModel: LogModel = viewModel(),
) {
    val listState = rememberLazyListState()
    val logLines by logModel.log.collectAsState()
    val scrollToBottom = remember(logLines.size) {
        logLines.isNotEmpty() && listState.layoutInfo.run {
            visibleItemsInfo.lastOrNull().let { info ->
                info == null || info.index == totalItemsCount - 1
            }
        }
    }

    if (scrollToBottom) LaunchedEffect(logLines.size) {
        listState.animateScrollToItem(logLines.size - 1)
    }

    LazyColumn(
        modifier = modifier.fillMaxWidth(),
        state = listState
    ) {
        items(logLines) { line ->
            Text(
                text = line,
                modifier = Modifier.padding(3.dp),
            )
        }
    }
}
