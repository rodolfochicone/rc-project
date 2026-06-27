import { type ReactElement, type ReactNode } from "react";

import {
  Alert,
  EmptyState,
  SectionHeading,
  SkeletonRow,
  StatusBadge,
  SurfaceCard,
  SurfaceCardBody,
  SurfaceCardDescription,
  SurfaceCardEyebrow,
  SurfaceCardHeader,
  SurfaceCardTitle,
} from "@rodolfochicone/ui";
import { Blocks, Bot } from "lucide-react";

import { apiErrorMessage } from "@/lib/api-client";

import { useCatalogAgents, useCatalogExtensions } from "../hooks/use-extensions";
import type { AgentItem, ExtensionItem } from "../types";

export interface ExtensionsViewProps {
  workspaceId: string;
}

export function ExtensionsView({ workspaceId }: ExtensionsViewProps): ReactElement {
  const agentsQuery = useCatalogAgents(workspaceId);
  const extensionsQuery = useCatalogExtensions(workspaceId);

  const extensions = extensionsQuery.data ?? [];
  const agents = agentsQuery.data ?? [];

  return (
    <div className="space-y-8" data-testid="extensions-view">
      <SectionHeading
        description="Installed extensions and reusable agents available to this workspace."
        eyebrow="Catalog"
        title="Extensions & agents"
      />

      <CatalogSection
        count={extensions.length}
        emptyDescription="No extensions are installed for this workspace yet."
        emptyIcon={<Blocks className="size-4" aria-hidden />}
        emptyTitle="No extensions"
        error={
          extensionsQuery.isError
            ? apiErrorMessage(extensionsQuery.error, "Failed to load extensions")
            : null
        }
        isLoading={extensionsQuery.isLoading}
        label="Extensions"
        testId="extensions-section"
      >
        <ul className="grid gap-3" data-testid="extensions-list">
          {extensions.map(ext => (
            <ExtensionRow ext={ext} key={ext.name} />
          ))}
        </ul>
      </CatalogSection>

      <CatalogSection
        count={agents.length}
        emptyDescription="No reusable agents are registered for this workspace yet."
        emptyIcon={<Bot className="size-4" aria-hidden />}
        emptyTitle="No agents"
        error={
          agentsQuery.isError ? apiErrorMessage(agentsQuery.error, "Failed to load agents") : null
        }
        isLoading={agentsQuery.isLoading}
        label="Reusable agents"
        testId="agents-section"
      >
        <ul className="grid gap-3" data-testid="agents-list">
          {agents.map(agent => (
            <AgentRow agent={agent} key={agent.name} />
          ))}
        </ul>
      </CatalogSection>
    </div>
  );
}

interface CatalogSectionProps {
  label: string;
  count: number;
  isLoading: boolean;
  error: string | null;
  testId: string;
  emptyTitle: string;
  emptyDescription: string;
  emptyIcon: ReactNode;
  children: ReactNode;
}

function CatalogSection({
  label,
  count,
  isLoading,
  error,
  testId,
  emptyTitle,
  emptyDescription,
  emptyIcon,
  children,
}: CatalogSectionProps): ReactElement {
  return (
    <section className="space-y-3" data-testid={testId}>
      <p className="eyebrow text-muted-foreground">
        {label}
        {!isLoading && !error ? ` · ${count}` : ""}
      </p>

      {error ? (
        <Alert data-testid={`${testId}-error`} variant="error">
          {error}
        </Alert>
      ) : isLoading ? (
        <div className="space-y-2" data-testid={`${testId}-loading`}>
          <SkeletonRow />
          <SkeletonRow />
        </div>
      ) : count === 0 ? (
        <EmptyState
          data-testid={`${testId}-empty`}
          description={emptyDescription}
          icon={emptyIcon}
          title={emptyTitle}
        />
      ) : (
        children
      )}
    </section>
  );
}

function ExtensionRow({ ext }: { ext: ExtensionItem }): ReactElement {
  return (
    <li>
      <SurfaceCard data-testid={`extension-item-${ext.name}`}>
        <SurfaceCardHeader>
          <div className="min-w-0">
            <SurfaceCardEyebrow>Extension</SurfaceCardEyebrow>
            <SurfaceCardTitle className="font-mono" data-testid={`extension-name-${ext.name}`}>
              {ext.name}
              {ext.version ? (
                <span className="ml-2 text-xs font-normal text-muted-foreground">
                  v{ext.version}
                </span>
              ) : null}
            </SurfaceCardTitle>
            <SurfaceCardDescription className="truncate font-mono text-xs" title={ext.source}>
              {ext.source}
            </SurfaceCardDescription>
          </div>
          <StatusBadge tone={ext.enabled ? "success" : "neutral"}>
            {ext.enabled ? "enabled" : "disabled"}
          </StatusBadge>
        </SurfaceCardHeader>
        {ext.description ? (
          <SurfaceCardBody data-testid={`extension-desc-${ext.name}`}>
            {ext.description}
          </SurfaceCardBody>
        ) : null}
      </SurfaceCard>
    </li>
  );
}

function AgentRow({ agent }: { agent: AgentItem }): ReactElement {
  const warnings = agent.warnings ?? [];
  return (
    <li>
      <SurfaceCard data-testid={`agent-item-${agent.name}`}>
        <SurfaceCardHeader>
          <div className="min-w-0">
            <SurfaceCardEyebrow>Agent</SurfaceCardEyebrow>
            <SurfaceCardTitle className="font-mono" data-testid={`agent-name-${agent.name}`}>
              {agent.name}
            </SurfaceCardTitle>
            {agent.description ? (
              <SurfaceCardDescription data-testid={`agent-desc-${agent.name}`}>
                {agent.description}
              </SurfaceCardDescription>
            ) : null}
          </div>
          <StatusBadge tone="info">{agent.scope}</StatusBadge>
        </SurfaceCardHeader>
        {warnings.length > 0 ? (
          <SurfaceCardBody className="space-y-2">
            {warnings.map((warning, index) => (
              <Alert
                data-testid={`agent-warning-${agent.name}-${index}`}
                key={warning}
                variant="warning"
              >
                {warning}
              </Alert>
            ))}
          </SurfaceCardBody>
        ) : null}
      </SurfaceCard>
    </li>
  );
}
