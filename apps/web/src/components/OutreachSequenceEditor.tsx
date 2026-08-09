"use client";

import { useEffect, useMemo, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import { formatDate } from "@/lib/constants";
import {
  countTemplateLinks,
  createBlankStep,
  renderLocalTemplate,
  validateSequence,
  validateSequenceStep,
} from "@/lib/outreach-sequence";
import type {
  OutreachSequence,
  OutreachSequencePreview,
  OutreachSequenceStep,
} from "@/lib/types";
import { EmptyState, ErrorBanner, StatusBadge } from "@/components/ui";

type Props = {
  sequences: OutreachSequence[];
  activeSequenceId?: string;
  loading: boolean;
  onReload: () => Promise<void>;
};

function cloneSteps(steps: OutreachSequenceStep[]): OutreachSequenceStep[] {
  return [...steps]
    .sort((left, right) => left.position - right.position)
    .map((step, index) => ({ ...step, position: index + 1 }));
}

function normalizeSequence(value: unknown): OutreachSequence | null {
  if (!value || typeof value !== "object") return null;
  if ("sequence" in value) {
    return (value as { sequence?: OutreachSequence }).sequence || null;
  }
  return value as OutreachSequence;
}

export function OutreachSequenceEditor({
  sequences,
  activeSequenceId,
  loading,
  onReload,
}: Props) {
  const [selectedId, setSelectedId] = useState("");
  const [pendingSelectionId, setPendingSelectionId] = useState("");
  const [name, setName] = useState("");
  const [steps, setSteps] = useState<OutreachSequenceStep[]>([]);
  const [dirty, setDirty] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [sampleRestaurant, setSampleRestaurant] = useState("Sample Restaurant");
  const [sampleOwner, setSampleOwner] = useState("");
  const [previewPosition, setPreviewPosition] = useState(1);
  const [serverPreview, setServerPreview] =
    useState<OutreachSequencePreview | null>(null);

  const selected = useMemo(
    () => sequences.find((sequence) => sequence.id === selectedId) || null,
    [selectedId, sequences],
  );
  const editable = selected?.status === "draft";
  const issues = useMemo(() => validateSequence(steps), [steps]);
  const selectedPreviewStep =
    steps.find((step) => step.position === previewPosition) || steps[0];
  const serverPreviewStep = serverPreview?.steps.find(
    (step) => step.position === selectedPreviewStep?.position,
  );
  const firstEnabledPosition = steps.find((step) => step.enabled)?.position;

  useEffect(() => {
    if (sequences.length === 0) {
      setSelectedId("");
      setName("");
      setSteps([]);
      return;
    }
    if (pendingSelectionId) {
      if (sequences.some((sequence) => sequence.id === pendingSelectionId)) {
        setSelectedId(pendingSelectionId);
        setPendingSelectionId("");
      }
      return;
    }
    if (!sequences.some((sequence) => sequence.id === selectedId)) {
      const initial =
        sequences.find((sequence) => sequence.status === "draft") ||
        sequences.find((sequence) => sequence.id === activeSequenceId) ||
        sequences[0];
      setSelectedId(initial.id);
    }
  }, [activeSequenceId, pendingSelectionId, selectedId, sequences]);

  useEffect(() => {
    if (!selected) return;
    setName(selected.name);
    setSteps(cloneSteps(selected.steps || []));
    setPreviewPosition(selected.steps?.[0]?.position || 1);
    setDirty(false);
    setServerPreview(null);
    setError(null);
  }, [selected]);

  function chooseSequence(id: string) {
    if (dirty && !confirm("Discard the unsaved sequence edits?")) return;
    setSelectedId(id);
  }

  function updateStep(
    index: number,
    patch: Partial<OutreachSequenceStep>,
  ) {
    setSteps((current) =>
      current.map((step, stepIndex) =>
        stepIndex === index ? { ...step, ...patch } : step,
      ),
    );
    setDirty(true);
    setServerPreview(null);
    setMessage(null);
  }

  function moveStep(index: number, direction: -1 | 1) {
    const target = index + direction;
    if (target < 0 || target >= steps.length) return;
    setSteps((current) => {
      const next = [...current];
      [next[index], next[target]] = [next[target], next[index]];
      return next.map((step, stepIndex) => ({
        ...step,
        position: stepIndex + 1,
      }));
    });
    setPreviewPosition(target + 1);
    setDirty(true);
    setServerPreview(null);
  }

  function removeStep(index: number) {
    if (steps.length <= 1) return;
    if (!confirm(`Remove email ${index + 1} from this draft?`)) return;
    const next = steps
      .filter((_, stepIndex) => stepIndex !== index)
      .map((step, stepIndex) => ({ ...step, position: stepIndex + 1 }));
    setSteps(next);
    setPreviewPosition(Math.min(previewPosition, next.length));
    setDirty(true);
    setServerPreview(null);
  }

  function addStep() {
    const position = steps.length + 1;
    setSteps((current) => [...current, createBlankStep(position)]);
    setPreviewPosition(position);
    setDirty(true);
    setServerPreview(null);
  }

  async function createDraft() {
    setBusy(true);
    setError(null);
    setMessage(null);
    try {
      const basedOn = selected || sequences.find((item) => item.is_active);
      const result = await adminFetch<OutreachSequence | { sequence: OutreachSequence }>(
        "outreach/sequences",
        {
          method: "POST",
          body: {
            name: basedOn?.name || "Restaurant outreach",
            based_on_sequence_id: basedOn?.id || undefined,
          },
        },
      );
      const created = normalizeSequence(result);
      if (created?.id) {
        setPendingSelectionId(created.id);
        setSelectedId(created.id);
      }
      await onReload();
      setMessage(
        basedOn
          ? "Editable draft version created. The active version remains unchanged until approval."
          : "Draft sequence created.",
      );
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not create draft");
    } finally {
      setBusy(false);
    }
  }

  async function saveDraft() {
    if (!selected || !editable) return;
    if (issues.length > 0) {
      setError("Resolve the validation issues before saving this draft.");
      return;
    }
    setBusy(true);
    setError(null);
    setMessage(null);
    try {
      await adminFetch(`outreach/sequences/${selected.id}`, {
        method: "PUT",
        body: {
          name: name.trim(),
          expected_updated_at: selected.updated_at,
          steps: steps.map((step, index) => ({
            ...step,
            position: index + 1,
          })),
        },
      });
      await onReload();
      setDirty(false);
      setMessage("Draft saved. Review the personalized preview before approval.");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not save draft");
    } finally {
      setBusy(false);
    }
  }

  async function approveDraft() {
    if (!selected || !editable || dirty || issues.length > 0) return;
    if (
      !confirm(
        `Approve version ${selected.version} and make it the active outreach sequence? The previous active version will be archived.`,
      )
    ) {
      return;
    }
    setBusy(true);
    setError(null);
    setMessage(null);
    try {
      await adminFetch(`outreach/sequences/${selected.id}/approve`, {
        method: "POST",
        body: { expected_updated_at: selected.updated_at },
      });
      await onReload();
      setMessage(
        "Sequence approved and made active. This does not enable the email job.",
      );
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not approve draft");
    } finally {
      setBusy(false);
    }
  }

  async function verifyPreview() {
    if (!selected) return;
    if (dirty) {
      setError("Save the draft before requesting a server-verified preview.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const result = await adminFetch<OutreachSequencePreview>(
        `outreach/sequences/${selected.id}/preview`,
        {
          method: "POST",
          body: {
            restaurant_name: sampleRestaurant.trim() || "Sample Restaurant",
            owner_first_name: sampleOwner.trim() || undefined,
          },
        },
      );
      setServerPreview(result);
      setMessage("Preview verified by the server from the saved sequence.");
    } catch (reason) {
      setError(reason instanceof Error ? reason.message : "Could not render preview");
    } finally {
      setBusy(false);
    }
  }

  if (loading && sequences.length === 0) {
    return <EmptyState message="Loading outreach sequence…" />;
  }

  if (sequences.length === 0) {
    return (
      <div className="card sequence-empty">
        <h2>No sequence exists yet</h2>
        <p>
          Create a draft, add the approved plain-text emails, then explicitly
          approve the version before the email job can use it.
        </p>
        <button className="btn btn-primary" type="button" onClick={createDraft} disabled={busy}>
          {busy ? "Creating…" : "Create first draft"}
        </button>
        <ErrorBanner message={error} />
      </div>
    );
  }

  return (
    <div className="sequence-workspace">
      <div className="card sequence-toolbar">
        <div className="sequence-toolbar-main">
          <label className="field-label" htmlFor="sequence-version">
            Sequence version
          </label>
          <select
            id="sequence-version"
            className="select"
            value={selectedId}
            onChange={(event) => chooseSequence(event.target.value)}
          >
            {sequences.map((sequence) => (
              <option key={sequence.id} value={sequence.id}>
                v{sequence.version} · {sequence.name} · {sequence.is_active ? "active" : sequence.status}
              </option>
            ))}
          </select>
        </div>
        <div className="sequence-toolbar-actions">
          {selected ? <StatusBadge status={selected.is_active ? "active" : selected.status} /> : null}
          <button className="btn btn-secondary" type="button" onClick={createDraft} disabled={busy}>
            {editable ? "Create another draft" : "Create editable version"}
          </button>
        </div>
        {selected ? (
          <p className="sequence-meta">
            Version {selected.version} · updated {formatDate(selected.updated_at)}
            {selected.approved_at ? ` · approved ${formatDate(selected.approved_at)}` : ""}
          </p>
        ) : null}
      </div>

      <ErrorBanner message={error} />
      {message ? (
        <div className="alert alert-info" role="status" aria-live="polite">
          {message}
        </div>
      ) : null}

      {!editable ? (
        <div className="alert alert-info">
          Approved and archived versions are read-only. Create an editable version to make changes;
          the currently active sequence keeps running until the new draft is approved.
        </div>
      ) : null}

      <div className="sequence-columns">
        <section className="sequence-editor" aria-labelledby="sequence-editor-title">
          <div className="card sequence-name-card">
            <div>
              <h2 id="sequence-editor-title">Email sequence</h2>
              <p>
                Enabled emails send in order. Delay is measured from the previous confirmed send;
                failures never advance a recipient.
              </p>
            </div>
            <label className="field-label" htmlFor="sequence-name">
              Internal sequence name
              <input
                id="sequence-name"
                className="input"
                value={name}
                onChange={(event) => {
                  setName(event.target.value);
                  setDirty(true);
                  setMessage(null);
                }}
                disabled={!editable}
                maxLength={120}
                required
              />
            </label>
          </div>

          <div className="sequence-step-list">
            {steps.map((step, index) => {
              const stepIssues = validateSequenceStep(
                step,
                step.position === firstEnabledPosition ? 0 : 1,
              );
              return (
                <fieldset
                  className="card sequence-step"
                  key={step.id || `new-${index}`}
                  disabled={!editable}
                >
                  <legend>Email {index + 1}</legend>
                  <div className="sequence-step-heading">
                    <label className="toggle-label">
                      <input
                        type="checkbox"
                        checked={step.enabled}
                        onChange={(event) => updateStep(index, { enabled: event.target.checked })}
                      />
                      <span>{step.enabled ? "Enabled" : "Disabled"}</span>
                    </label>
                    <div className="sequence-order-actions" aria-label={`Reorder email ${index + 1}`}>
                      <button
                        className="btn btn-secondary btn-compact"
                        type="button"
                        onClick={() => moveStep(index, -1)}
                        disabled={!editable || index === 0}
                        aria-label={`Move email ${index + 1} up`}
                      >
                        Move up
                      </button>
                      <button
                        className="btn btn-secondary btn-compact"
                        type="button"
                        onClick={() => moveStep(index, 1)}
                        disabled={!editable || index === steps.length - 1}
                        aria-label={`Move email ${index + 1} down`}
                      >
                        Move down
                      </button>
                      <button
                        className="btn btn-danger btn-compact"
                        type="button"
                        onClick={() => removeStep(index)}
                        disabled={!editable || steps.length <= 1}
                      >
                        Remove
                      </button>
                    </div>
                  </div>

                  <label className="field-label" htmlFor={`step-${index}-delay`}>
                    {index === 0 ? "Initial delay (days)" : "Wait after previous confirmed send (days)"}
                    <input
                      id={`step-${index}-delay`}
                      className="input delay-input"
                      type="number"
                      min="0"
                      step="0.25"
                      inputMode="decimal"
                      value={step.delay_hours / 24}
                      onChange={(event) =>
                        updateStep(index, {
                          delay_hours: Math.max(0, Math.round(Number(event.target.value || 0) * 24)),
                        })
                      }
                    />
                    <span className="field-help">
                      Stored as {step.delay_hours} hours. The default follow-up interval is 3 days (72 hours).
                    </span>
                  </label>

                  <label className="field-label" htmlFor={`step-${index}-subject`}>
                    Subject
                    <input
                      id={`step-${index}-subject`}
                      className="input"
                      value={step.subject_template}
                      onChange={(event) => updateStep(index, { subject_template: event.target.value })}
                      maxLength={200}
                      spellCheck
                    />
                  </label>

                  <label className="field-label" htmlFor={`step-${index}-body`}>
                    Plain-text message
                    <textarea
                      id={`step-${index}-body`}
                      className="textarea sequence-body"
                      value={step.body_text_template}
                      onChange={(event) => updateStep(index, { body_text_template: event.target.value })}
                      rows={14}
                      spellCheck
                    />
                  </label>

                  <div className="sequence-link-check" data-valid={stepIssues.length === 0}>
                    <strong>{countTemplateLinks(step)}/2 links</strong>
                    <span>
                      Use <code>{"{{website_url}}"}</code> and <code>{"{{unsubscribe_url}}"}</code> exactly once each.
                    </span>
                  </div>
                  {stepIssues.length > 0 ? (
                    <ul className="sequence-issues" aria-label={`Email ${index + 1} validation issues`}>
                      {stepIssues.map((issue) => (
                        <li key={issue}>{issue}</li>
                      ))}
                    </ul>
                  ) : null}
                </fieldset>
              );
            })}
          </div>

          <div className="sequence-draft-actions">
            <button className="btn btn-secondary" type="button" onClick={addStep} disabled={!editable || busy}>
              Add email
            </button>
            <span className="sequence-save-state" role="status">
              {dirty ? "Unsaved changes" : "Saved"}
            </span>
            <button
              className="btn btn-primary"
              type="button"
              onClick={saveDraft}
              disabled={!editable || busy || !dirty || issues.length > 0 || !name.trim()}
            >
              {busy ? "Working…" : "Save draft"}
            </button>
            <button
              className="btn btn-primary"
              type="button"
              onClick={approveDraft}
              disabled={!editable || busy || dirty || issues.length > 0}
            >
              Approve &amp; make active
            </button>
          </div>
          {issues.length > 0 ? (
            <div className="alert alert-error" role="alert">
              <strong>{issues.length} issue{issues.length === 1 ? "" : "s"} to resolve</strong>
              <ul className="sequence-issues">
                {issues.map((issue) => (
                  <li key={issue}>{issue}</li>
                ))}
              </ul>
            </div>
          ) : null}
        </section>

        <aside className="card sequence-preview" aria-labelledby="sequence-preview-title">
          <div>
            <h2 id="sequence-preview-title">Personalized preview</h2>
            <p>Owner first name is used when available; otherwise the restaurant team is addressed.</p>
          </div>
          <label className="field-label" htmlFor="preview-restaurant">
            Restaurant
            <input
              id="preview-restaurant"
              className="input"
              value={sampleRestaurant}
              onChange={(event) => {
                setSampleRestaurant(event.target.value);
                setServerPreview(null);
              }}
            />
          </label>
          <label className="field-label" htmlFor="preview-owner">
            Owner first name (optional)
            <input
              id="preview-owner"
              className="input"
              value={sampleOwner}
              onChange={(event) => {
                setSampleOwner(event.target.value);
                setServerPreview(null);
              }}
            />
          </label>
          <label className="field-label" htmlFor="preview-email">
            Email to preview
            <select
              id="preview-email"
              className="select"
              value={selectedPreviewStep?.position || 1}
              onChange={(event) => setPreviewPosition(Number(event.target.value))}
            >
              {steps.map((step) => (
                <option key={step.position} value={step.position}>
                  Email {step.position}{step.enabled ? "" : " (disabled)"}
                </option>
              ))}
            </select>
          </label>
          <button
            className="btn btn-secondary"
            type="button"
            onClick={verifyPreview}
            disabled={!selected || busy || dirty}
          >
            {busy ? "Rendering…" : "Verify saved preview"}
          </button>

          {selectedPreviewStep ? (
            <div className="email-preview" aria-live="polite">
              <div className="email-preview-subject">
                <span>Subject</span>
                <strong>
                  {serverPreviewStep?.subject ||
                    renderLocalTemplate(
                      selectedPreviewStep.subject_template,
                      sampleRestaurant,
                      sampleOwner,
                    )}
                </strong>
              </div>
              <pre>
                {serverPreviewStep?.body_text ||
                  renderLocalTemplate(
                    selectedPreviewStep.body_text_template,
                    sampleRestaurant,
                    sampleOwner,
                  )}
              </pre>
              <div className="sequence-link-check" data-valid={(serverPreviewStep?.url_count ?? countTemplateLinks(selectedPreviewStep)) === 2}>
                <strong>{serverPreviewStep?.url_count ?? countTemplateLinks(selectedPreviewStep)}/2 links</strong>
                <span>{serverPreviewStep ? "Server-verified saved preview" : "Live preview of current editor values"}</span>
              </div>
            </div>
          ) : (
            <EmptyState message="Add an email to preview it." />
          )}
        </aside>
      </div>
    </div>
  );
}
