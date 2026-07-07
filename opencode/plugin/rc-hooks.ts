import type { Plugin } from "@opencode-ai/plugin";
import { join } from "node:path";

// rc hook parity for OpenCode. OpenCode has no settings.json hooks like Claude
// Code; it exposes a plugin API instead. Rather than reimplement the guard logic
// in TypeScript, this plugin shells out to the SAME bundled shell scripts the
// Claude channel uses (git-guard, commit-guard, go-mod-guard, gateguard, go-fmt),
// so there is one source of truth. When installing manually, copy this repo's
// hooks/scripts/ so they are reachable at <opencode-root>/rc/hooks/scripts/.
//
// The scripts read a Claude-style JSON payload on stdin and exit 2 to block; this
// plugin maps OpenCode's tool events to that payload and turns an exit-2 into a
// thrown error (which OpenCode treats as a denied tool call). The same env
// knobs apply: RC_HOOK_PROFILE, RC_DISABLED_HOOKS, RC_DRY_RUN.

const SCRIPTS_DIR = join(import.meta.dir, "..", "rc", "hooks", "scripts");

export const RcHooks: Plugin = async ({ $ }) => {
  // This plugin may be installed into BOTH .opencode/plugin/ and
  // .opencode/plugins/ for cross-version compatibility (opencode has used both
  // names). If a version loads both copies in the same process, register the
  // hooks only once so guards don't run twice.
  const g = globalThis as Record<string, unknown>;
  if (g.__rcHooksActive) return {};
  g.__rcHooksActive = true;

  // Run one guard script with a synthesized payload. Exit 2 → throw (block).
  // Any other failure (missing bash/jq, exit 0/1) is non-blocking, matching the
  // scripts' fail-open contract so a broken environment never wedges a session.
  async function runGuard(script: string, payload: unknown): Promise<void> {
    const scriptPath = join(SCRIPTS_DIR, script);
    const res = await $`printf '%s' ${JSON.stringify(payload)} | bash ${scriptPath}`
      .nothrow()
      .quiet();
    if (res.exitCode === 2) {
      const msg = res.stderr.toString().trim() || `blocked by rc ${script}`;
      throw new Error(msg);
    }
  }

  function fileArgs(input: any, output: any) {
    const filePath = output?.args?.filePath ?? input?.args?.filePath ?? "";
    return { tool_input: { file_path: filePath }, session_id: input?.sessionID ?? "opencode" };
  }

  // Notification sound — opt-in via RC_SOUND=1. Fire-and-forget so it never
  // delays the turn; afplay on macOS, ignored on platforms without it.
  function playSound(kind: "done" | "attention") {
    if (process.env.RC_SOUND !== "1") return;
    const sound =
      kind === "attention"
        ? "/System/Library/Sounds/Funk.aiff"
        : "/System/Library/Sounds/Hero.aiff";
    void $`afplay ${sound}`.nothrow().quiet();
  }

  return {
    "tool.execute.before": async (input: any, output: any) => {
      if (input.tool === "bash") {
        const payload = {
          tool_input: { command: output?.args?.command ?? input?.args?.command ?? "" },
        };
        await runGuard("git-guard.sh", payload);
        await runGuard("commit-guard.sh", payload);
        return;
      }
      if (input.tool === "edit" || input.tool === "write" || input.tool === "patch") {
        const payload = fileArgs(input, output);
        await runGuard("go-mod-guard.sh", payload);
        await runGuard("gateguard.sh", payload);
      }
    },
    "tool.execute.after": async (input: any, output: any) => {
      const isEdit = input.tool === "edit" || input.tool === "write" || input.tool === "patch";
      if (isEdit) {
        // go-fmt never blocks (exit 0); it normalizes the edited Go file.
        await runGuard("go-fmt.sh", fileArgs(input, output));
      }
      // Instincts capture — opt-in, only when RC_INSTINCTS=1 (observe.sh self-gates
      // too, but skip spawning bash entirely when disabled).
      if (process.env.RC_INSTINCTS === "1" && (isEdit || input.tool === "bash")) {
        const args = output?.args ?? input?.args ?? {};
        const payload =
          input.tool === "bash"
            ? { tool_name: "bash", tool_input: { command: args.command ?? "" } }
            : { tool_name: input.tool, tool_input: { file_path: args.filePath ?? "" } };
        await runGuard("observe.sh", payload);
      }
    },
    // End-of-turn / needs-attention notification sound (opt-in via RC_SOUND=1).
    event: async ({ event }: any) => {
      const type = event?.type;
      if (type === "session.idle") playSound("done");
      else if (type === "permission.asked") playSound("attention");
    },
  };
};
