_astral_query_cache_spec() {
    local target="$1"
    local cache_var="_astral_query_spec_cache_${target:-local}"
    cache_var="${cache_var//[^a-zA-Z0-9_]/_}"

    if [[ -z "${!cache_var}" ]]; then
        local cmd="astral-query"
        local op=".spec"
        if [[ -n "$target" ]]; then
            op="${target}:.spec"
        fi
        local spec
        spec=$($cmd "$op" -out json 2>/dev/null | jq -c 'select(.Type != "eos")')
        if [[ $? -ne 0 || -z "$spec" ]]; then
            return 1
        fi
        printf -v "$cache_var" '%s' "$spec"
    fi
    echo "${!cache_var}"
}

_astral_query_cache_aliases() {
    if [[ -z "$_astral_query_alias_cache" ]]; then
        local result
        result=$(astral-query dir.alias_map -out json 2>/dev/null) || return 1
        _astral_query_alias_cache="$result"
    fi
    echo "$_astral_query_alias_cache"
}

_astral_query_get_param_type() {
    local spec="$1" op="$2" param="$3"
    echo "$spec" | jq -r --arg op "$op" --arg param "$param" \
        'select(.Object.Name == $op) | .Object.Parameters[] | select(.Name == $param) | .Type // empty' 2>/dev/null
}

_astral_query_completions() {
    local cur prev words cword
    _init_completion -n : || return

    # Parse the command line to find target and operation
    local target="" op="" op_index=0
    local i
    for (( i=1; i < cword; i++ )); do
        local word="${words[$i]}"
        # Skip flags and their values
        if [[ "$word" == -* ]]; then
            continue
        fi
        # First non-flag argument is [target:]operation
        if [[ $op_index -eq 0 ]]; then
            if [[ "$word" == *:* ]]; then
                target="${word%%:*}"
                op="${word#*:}"
            else
                op="$word"
            fi
            op_index=$i
            break
        fi
    done

    # If we're completing the first positional arg (operation name)
    if [[ $op_index -eq 0 ]] || [[ $cword -eq $op_index ]]; then
        # Check if the current word has a target: prefix
        if [[ "$cur" == *:* ]]; then
            local prefix="${cur%%:*}:"
            local spec
            spec=$(_astral_query_cache_spec "${cur%%:*}") || return
            local ops
            ops=$(echo "$spec" | jq -r '.Object.Name' 2>/dev/null)
            local completions=()
            while IFS= read -r name; do
                [[ -z "$name" ]] && continue
                completions+=("${prefix}${name}")
            done <<< "$ops"
            COMPREPLY=($(compgen -W "${completions[*]}" -- "$cur"))
        else
            local spec
            spec=$(_astral_query_cache_spec "") || return
            local ops
            ops=$(echo "$spec" | jq -r '.Object.Name' 2>/dev/null)
            COMPREPLY=($(compgen -W "$ops" -- "$cur"))
        fi
        __ltrim_colon_completions "$cur"
        return
    fi

    local spec
    spec=$(_astral_query_cache_spec "$target") || return

    # We're past the operation name — complete parameter names
    if [[ "$cur" == -* ]]; then
        local params
        params=$(echo "$spec" | jq -r --arg op "$op" \
            'select(.Object.Name == $op) | .Object.Parameters[].Name' 2>/dev/null)

        local completions=()
        while IFS= read -r param; do
            [[ -z "$param" ]] && continue
            completions+=("-${param}")
        done <<< "$params"
        COMPREPLY=($(compgen -W "${completions[*]}" -- "$cur"))
        __ltrim_colon_completions "$cur"
        return
    fi

    # Completing a parameter value — check if previous word is a flag
    if [[ "$prev" == -* ]]; then
        local param_name="${prev#-}"

        local param_type
        param_type=$(_astral_query_get_param_type "$spec" "$op" "$param_name")

        if [[ "$param_type" == "identity" ]]; then
            local alias_data
            alias_data=$(_astral_query_cache_aliases) || return
            local identities
            identities=$(echo "$alias_data" | jq -r \
                '.Object.Aliases | to_entries[] | .key, .value' 2>/dev/null)
            COMPREPLY=($(compgen -W "$identities" -- "$cur"))
            __ltrim_colon_completions "$cur"
            return
        fi
    fi
}

complete -F _astral_query_completions astral-query
