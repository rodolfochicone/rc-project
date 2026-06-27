# rc Completion Plugin

This plugin adds shell completion for `rc tasks run` so task slugs are completed from the
nearest `.rc/tasks` directory relative to your current working directory.

## What it does

- Completes `tasks` after `rc`
- Completes `run` after `rc tasks`
- Completes all directories under `.rc/tasks` after `rc tasks run`
- Works in worktrees and repository copies by scanning upward from `$PWD` until it finds `.rc/tasks`

## Installation

1. Copy the plugin file into your shell folder (already placed by default at:
   `~/.zsh/rc-completion/rc-completion.plugin.zsh`).

   ```zsh
   # if needed
   cp /path/to/rc/zsh/rc-completion/rc-completion.plugin.zsh \
     "$HOME/.zsh/rc-completion/rc-completion.plugin.zsh"
   ```

2. Source it from your `~/.zshrc`:

   ```zsh
   if [[ -f "$HOME/.zsh/rc-completion/rc-completion.plugin.zsh" ]]; then
     source "$HOME/.zsh/rc-completion/rc-completion.plugin.zsh"
   fi
   ```

3. Reload your shell:

   ```zsh
   source ~/.zshrc
   ```

## Quick usage

From any rc workspace:

```zsh
cd /path/to/repo/.rc-task-root
rc tasks run <TAB>
```

The command will suggest task directory names found in `.rc/tasks`.

## Notes

- Keep `.rc/tasks` present in the workspace root or an ancestor directory.
- If there are no tasks, completion will fall back to default zsh behavior for that command position.
