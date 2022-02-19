package cc.cryptopunks.astral.service.ui

import android.view.ViewGroup
import android.widget.TextView
import androidx.core.view.setPadding
import androidx.recyclerview.widget.RecyclerView
import cc.cryptopunks.astral.service.R

class LogAdapter : RecyclerView.Adapter<LogAdapter.ViewHolder>() {

    val lines: MutableList<String> = mutableListOf()

    override fun getItemCount(): Int = lines.size

    override fun onCreateViewHolder(parent: ViewGroup, viewType: Int): ViewHolder {
        val textView = TextView(parent.context).apply {
            setPadding(resources.getDimensionPixelSize(R.dimen.log_padding))
        }
        return ViewHolder(textView)
    }

    override fun onBindViewHolder(holder: ViewHolder, position: Int) {
        holder.view.text = lines[position]
    }

    class ViewHolder(val view: TextView) : RecyclerView.ViewHolder(view)
}
