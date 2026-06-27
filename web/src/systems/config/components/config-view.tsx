import { useId, useMemo, useState, type ReactElement, type ReactNode } from "react";

import {
  Alert,
  Button,
  SectionHeading,
  SurfaceCard,
  SurfaceCardBody,
  SurfaceCardEyebrow,
  SurfaceCardHeader,
  SurfaceCardTitle,
} from "@rodolfochicone/ui";
import { Save } from "lucide-react";

import { apiErrorMessage } from "@/lib/api-client";
import { ACCESS_MODES, modelsForRuntime, REASONING_EFFORTS, RUNTIMES } from "@/lib/runtime-catalog";

import { useGlobalConfig, useSaveGlobalConfig } from "../hooks/use-config";
import type { ConfigDefaults, ConfigDocument, ConfigRuns, ConfigSound } from "../types";

const fieldClass =
  "w-full rounded-[var(--radius-md)] border border-border bg-[color:var(--surface-inset)] px-3 py-2 text-sm text-foreground transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/60";
const labelClass = "block text-sm font-medium text-foreground";

const RUNTIME_OPTIONS = [
  { value: "", label: "Runtime default" },
  ...RUNTIMES.map(runtime => ({ value: runtime.id, label: runtime.label })),
];
const REASONING_OPTIONS = [
  { value: "", label: "Runtime default" },
  ...REASONING_EFFORTS.map(effort => ({ value: effort, label: effort })),
];
const ACCESS_OPTIONS = [
  { value: "", label: "Runtime default" },
  ...ACCESS_MODES.map(mode => ({ value: mode, label: mode })),
];
const ATTACH_OPTIONS = [
  { value: "", label: "Config default" },
  { value: "auto", label: "auto" },
  { value: "ui", label: "ui" },
  { value: "stream", label: "stream" },
  { value: "detach", label: "detach" },
];

export interface ConfigViewProps {
  workspaceId: string;
}

export function ConfigView({ workspaceId: _workspaceId }: ConfigViewProps): ReactElement {
  const globalQuery = useGlobalConfig();
  const saveGlobal = useSaveGlobalConfig();
  const [saveError, setSaveError] = useState<string | null>(null);
  const [savedAt, setSavedAt] = useState<number | null>(null);

  function handleSave(doc: ConfigDocument): void {
    setSaveError(null);
    saveGlobal.mutate(doc, {
      onSuccess: () => setSavedAt(Date.now()),
      onError: err => setSaveError(apiErrorMessage(err, "Failed to save config")),
    });
  }

  if (globalQuery.isLoading) {
    return (
      <div className="text-sm text-muted-foreground" data-testid="config-loading">
        Loading configuration…
      </div>
    );
  }

  if (globalQuery.isError) {
    return (
      <Alert data-testid="config-error" variant="error">
        {apiErrorMessage(globalQuery.error, "Failed to load config")}
      </Alert>
    );
  }

  const config = globalQuery.data ?? {};

  return (
    <ConfigEditor
      // Re-seed the form from server state whenever it changes (e.g. after a save).
      key={JSON.stringify(config)}
      initial={config}
      isSaving={saveGlobal.isPending}
      onSave={handleSave}
      saveError={saveError}
      savedAt={savedAt}
    />
  );
}

interface ConfigEditorProps {
  initial: ConfigDocument;
  isSaving: boolean;
  saveError: string | null;
  savedAt: number | null;
  onSave: (doc: ConfigDocument) => void;
}

function ConfigEditor({
  initial,
  isSaving,
  saveError,
  savedAt,
  onSave,
}: ConfigEditorProps): ReactElement {
  const [defaults, setDefaults] = useState<ConfigDefaults>(initial.defaults ?? {});
  const [runs, setRuns] = useState<ConfigRuns>(initial.runs ?? {});
  const [sound, setSound] = useState<ConfigSound>(initial.sound ?? {});

  const dirty = useMemo(() => {
    const current = JSON.stringify({ defaults, runs, sound });
    const baseline = JSON.stringify({
      defaults: initial.defaults ?? {},
      runs: initial.runs ?? {},
      sound: initial.sound ?? {},
    });
    return current !== baseline;
  }, [defaults, runs, sound, initial]);

  const addDirsText = (defaults.add_dirs ?? []).join("\n");

  function submit(): void {
    onSave({ ...initial, defaults, runs, sound });
  }

  return (
    <div className="space-y-6" data-testid="config-view">
      <SectionHeading
        actions={
          <Button
            data-testid="config-save-btn"
            disabled={!dirty || isSaving}
            icon={<Save className="size-4" />}
            loading={isSaving}
            onClick={submit}
          >
            Save changes
          </Button>
        }
        description="Defaults the daemon applies to every workspace unless a workspace overrides them."
        eyebrow="Global configuration"
        title="Settings"
      />

      {saveError ? (
        <Alert data-testid="config-save-error" variant="error">
          {saveError}
        </Alert>
      ) : null}
      {savedAt && !dirty ? (
        <Alert data-testid="config-save-success" variant="success">
          Configuration saved.
        </Alert>
      ) : null}

      <ConfigCard eyebrow="Runtime" title="Runtime defaults">
        <div className="grid gap-4 sm:grid-cols-2">
          <SelectField
            label="Runtime (IDE)"
            onChange={value => setDefaults(d => ({ ...d, ide: value }))}
            options={RUNTIME_OPTIONS}
            value={defaults.ide ?? ""}
          />
          <ComboField
            label="Model"
            onChange={value => setDefaults(d => ({ ...d, model: value }))}
            placeholder="Runtime default"
            suggestions={modelsForRuntime(defaults.ide ?? "")}
            value={defaults.model}
          />
          <SelectField
            label="Reasoning effort"
            onChange={value => setDefaults(d => ({ ...d, reasoning_effort: value }))}
            options={REASONING_OPTIONS}
            value={defaults.reasoning_effort ?? ""}
          />
          <SelectField
            label="Access mode"
            onChange={value => setDefaults(d => ({ ...d, access_mode: value }))}
            options={ACCESS_OPTIONS}
            value={defaults.access_mode ?? ""}
          />
          <TextField
            label="Output format"
            onChange={value => setDefaults(d => ({ ...d, output_format: value }))}
            placeholder="Runtime default"
            value={defaults.output_format}
          />
          <TextField
            label="Activity timeout"
            mono
            onChange={value => setDefaults(d => ({ ...d, timeout: value }))}
            placeholder="e.g. 10m"
            value={defaults.timeout}
          />
          <NumberField
            label="Tail lines"
            onChange={value => setDefaults(d => ({ ...d, tail_lines: value }))}
            value={defaults.tail_lines}
          />
          <NumberField
            label="Max retries"
            onChange={value => setDefaults(d => ({ ...d, max_retries: value }))}
            value={defaults.max_retries}
          />
          <NumberField
            label="Retry backoff multiplier"
            onChange={value => setDefaults(d => ({ ...d, retry_backoff_multiplier: value }))}
            step="0.1"
            value={defaults.retry_backoff_multiplier}
          />
        </div>
        <div className="mt-4 space-y-1.5">
          <label className={labelClass} htmlFor="config-add-dirs">
            Additional writable directories{" "}
            <span className="text-muted-foreground">(one per line)</span>
          </label>
          <textarea
            className={`${fieldClass} min-h-[5rem] resize-y font-mono`}
            data-testid="config-add-dirs"
            id="config-add-dirs"
            onChange={event =>
              setDefaults(d => ({
                ...d,
                add_dirs: event.target.value
                  .split("\n")
                  .map(line => line.trim())
                  .filter(Boolean),
              }))
            }
            placeholder="/absolute/path"
            value={addDirsText}
          />
        </div>
        <div className="mt-4">
          <CheckboxField
            checked={defaults.auto_commit ?? false}
            hint="Append automatic commit instructions to runs."
            label="Auto-commit"
            onChange={value => setDefaults(d => ({ ...d, auto_commit: value }))}
          />
        </div>
      </ConfigCard>

      <ConfigCard eyebrow="Runs" title="Runs & retention">
        <div className="grid gap-4 sm:grid-cols-2">
          <SelectField
            label="Default attach mode"
            onChange={value => setRuns(r => ({ ...r, default_attach_mode: value }))}
            options={ATTACH_OPTIONS}
            value={runs.default_attach_mode ?? ""}
          />
          <TextField
            label="Shutdown drain timeout"
            mono
            onChange={value => setRuns(r => ({ ...r, shutdown_drain_timeout: value }))}
            placeholder="e.g. 30s"
            value={runs.shutdown_drain_timeout}
          />
          <NumberField
            label="Keep terminal runs (days)"
            onChange={value => setRuns(r => ({ ...r, keep_terminal_days: value }))}
            value={runs.keep_terminal_days}
          />
          <NumberField
            label="Max terminal runs"
            onChange={value => setRuns(r => ({ ...r, keep_max: value }))}
            value={runs.keep_max}
          />
        </div>
      </ConfigCard>

      <ConfigCard eyebrow="Notifications" title="Completion sounds">
        <CheckboxField
          checked={sound.enabled ?? false}
          hint="Play a sound when a run finishes."
          label="Enable sounds"
          onChange={value => setSound(s => ({ ...s, enabled: value }))}
        />
        <div className="mt-4 grid gap-4 sm:grid-cols-2">
          <TextField
            label="On completed"
            mono
            onChange={value => setSound(s => ({ ...s, on_completed: value }))}
            placeholder="Sound name or path"
            value={sound.on_completed}
          />
          <TextField
            label="On failed"
            mono
            onChange={value => setSound(s => ({ ...s, on_failed: value }))}
            placeholder="Sound name or path"
            value={sound.on_failed}
          />
        </div>
      </ConfigCard>
    </div>
  );
}

function ConfigCard({
  eyebrow,
  title,
  children,
}: {
  eyebrow: string;
  title: string;
  children: ReactNode;
}): ReactElement {
  return (
    <SurfaceCard>
      <SurfaceCardHeader>
        <div>
          <SurfaceCardEyebrow>{eyebrow}</SurfaceCardEyebrow>
          <SurfaceCardTitle>{title}</SurfaceCardTitle>
        </div>
      </SurfaceCardHeader>
      <SurfaceCardBody>{children}</SurfaceCardBody>
    </SurfaceCard>
  );
}

function TextField({
  label,
  value,
  onChange,
  placeholder,
  mono = false,
}: {
  label: string;
  value: string | undefined;
  onChange: (value: string) => void;
  placeholder?: string;
  mono?: boolean;
}): ReactElement {
  return (
    <label className="block space-y-1.5">
      <span className={labelClass}>{label}</span>
      <input
        className={mono ? `${fieldClass} font-mono` : fieldClass}
        onChange={event => onChange(event.target.value)}
        placeholder={placeholder}
        value={value ?? ""}
      />
    </label>
  );
}

function ComboField({
  label,
  value,
  onChange,
  suggestions,
  placeholder,
}: {
  label: string;
  value: string | undefined;
  onChange: (value: string) => void;
  suggestions: readonly string[];
  placeholder?: string;
}): ReactElement {
  const listId = useId();
  return (
    <label className="block space-y-1.5">
      <span className={labelClass}>{label}</span>
      <input
        className={`${fieldClass} font-mono`}
        list={listId}
        onChange={event => onChange(event.target.value)}
        placeholder={placeholder}
        value={value ?? ""}
      />
      <datalist id={listId}>
        {suggestions.map(suggestion => (
          <option key={suggestion} value={suggestion} />
        ))}
      </datalist>
    </label>
  );
}

function NumberField({
  label,
  value,
  onChange,
  step,
}: {
  label: string;
  value: number | undefined;
  onChange: (value: number | undefined) => void;
  step?: string;
}): ReactElement {
  return (
    <label className="block space-y-1.5">
      <span className={labelClass}>{label}</span>
      <input
        className={`${fieldClass} font-mono`}
        onChange={event =>
          onChange(event.target.value === "" ? undefined : Number(event.target.value))
        }
        step={step}
        type="number"
        value={value ?? ""}
      />
    </label>
  );
}

function SelectField({
  label,
  value,
  onChange,
  options,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  options: { value: string; label: string }[];
}): ReactElement {
  return (
    <label className="block space-y-1.5">
      <span className={labelClass}>{label}</span>
      <select className={fieldClass} onChange={event => onChange(event.target.value)} value={value}>
        {options.map(option => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}

function CheckboxField({
  label,
  checked,
  onChange,
  hint,
}: {
  label: string;
  checked: boolean;
  onChange: (value: boolean) => void;
  hint?: string;
}): ReactElement {
  return (
    <label className="flex items-start gap-3">
      <input
        checked={checked}
        className="mt-0.5 size-4 rounded border-border accent-[color:var(--primary)]"
        onChange={event => onChange(event.target.checked)}
        type="checkbox"
      />
      <span>
        <span className="block text-sm font-medium text-foreground">{label}</span>
        {hint ? <span className="block text-xs text-muted-foreground">{hint}</span> : null}
      </span>
    </label>
  );
}
