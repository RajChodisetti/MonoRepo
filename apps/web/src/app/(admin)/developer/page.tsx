"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { adminFetch } from "@/lib/client-api";
import { ErrorBanner, PageHeader } from "@/components/ui";

type SchemaColumn = {
  name: string;
  data_type: string;
  database_type: string;
  is_nullable: boolean;
  column_default?: string;
  ordinal_position: number;
};

type SchemaTable = {
  name: string;
  type: string;
  columns: SchemaColumn[];
};

type QueryResult = {
  columns: { name: string }[];
  rows: unknown[][];
  row_count: number;
  truncated: boolean;
  elapsed_ms: number;
  max_rows: number;
};

const SAMPLE_QUERIES = [
  {
    label: "Menu popularity",
    query: `SELECT
  lower(trim(mi.name)) AS item_name,
  count(DISTINCT m.restaurant_id) AS restaurant_count,
  count(*) AS menu_item_rows,
  string_agg(DISTINCT r.name, ', ' ORDER BY r.name) AS restaurants
FROM menu_items mi
JOIN menus m ON m.id = mi.menu_id
JOIN restaurants r ON r.id = m.restaurant_id
WHERE trim(mi.name) <> ''
GROUP BY lower(trim(mi.name))
ORDER BY restaurant_count DESC, menu_item_rows DESC, item_name
LIMIT 100`,
  },
  {
    label: "Stored counts",
    query: `SELECT 'restaurants' AS dataset, count(*) AS rows FROM restaurants
UNION ALL SELECT 'menus', count(*) FROM menus
UNION ALL SELECT 'menu_items', count(*) FROM menu_items
ORDER BY dataset`,
  },
  {
    label: "Recent restaurants",
    query: `SELECT
  r.id,
  r.name,
  r.email,
  r.status,
  r.email_send_count,
  r.last_email_sent_at,
  r.updated_at
FROM restaurants r
ORDER BY r.updated_at DESC
LIMIT 50`,
  },
];

function formatCell(value: unknown): string {
  if (value === null || value === undefined) return "NULL";
  if (typeof value === "object") return JSON.stringify(value);
  return String(value);
}

function InlineEmpty({ message }: { message: string }) {
  return (
    <div
      style={{
        border: "1px dashed var(--line)",
        color: "var(--muted)",
        padding: "1rem",
        textAlign: "center",
      }}
    >
      {message}
    </div>
  );
}

export default function DeveloperPage() {
  const [schema, setSchema] = useState<SchemaTable[]>([]);
  const [schemaFilter, setSchemaFilter] = useState("");
  const [query, setQuery] = useState(SAMPLE_QUERIES[0].query);
  const [result, setResult] = useState<QueryResult | null>(null);
  const [loadingSchema, setLoadingSchema] = useState(true);
  const [executing, setExecuting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadSchema = useCallback(async () => {
    setLoadingSchema(true);
    setError(null);
    try {
      const data = await adminFetch<{ tables: SchemaTable[] }>("developer/schema");
      setSchema(data.tables || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load schema");
    } finally {
      setLoadingSchema(false);
    }
  }, []);

  useEffect(() => {
    loadSchema();
  }, [loadSchema]);

  const filteredSchema = useMemo(() => {
    const needle = schemaFilter.trim().toLowerCase();
    if (!needle) return schema;
    return schema
      .map((table) => ({
        ...table,
        columns: table.columns.filter(
          (column) =>
            table.name.toLowerCase().includes(needle) ||
            column.name.toLowerCase().includes(needle) ||
            column.data_type.toLowerCase().includes(needle),
        ),
      }))
      .filter(
        (table) =>
          table.name.toLowerCase().includes(needle) || table.columns.length > 0,
      );
  }, [schema, schemaFilter]);

  async function executeQuery() {
    setExecuting(true);
    setError(null);
    setResult(null);
    try {
      const data = await adminFetch<QueryResult>("developer/sql", {
        method: "POST",
        body: { query },
      });
      setResult(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Query failed");
    } finally {
      setExecuting(false);
    }
  }

  return (
    <div>
      <PageHeader
        title="Developer"
        subtitle="Read-only PostgreSQL query console and schema browser"
        actions={
          <button className="btn btn-secondary" type="button" onClick={loadSchema}>
            Refresh schema
          </button>
        }
      />
      <ErrorBanner message={error} />

      <div className="card" style={{ marginBottom: "1rem", display: "grid", gap: "0.85rem" }}>
        <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
          {SAMPLE_QUERIES.map((sample) => (
            <button
              key={sample.label}
              className="btn btn-secondary"
              type="button"
              onClick={() => setQuery(sample.query)}
            >
              {sample.label}
            </button>
          ))}
        </div>
        <label style={{ display: "grid", gap: "0.35rem" }}>
          <span style={{ fontSize: "0.85rem", fontWeight: 600 }}>SQL</span>
          <textarea
            className="textarea"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            rows={11}
            spellCheck={false}
            style={{
              fontFamily: "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace",
              lineHeight: 1.45,
              resize: "vertical",
            }}
          />
        </label>
        <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap", alignItems: "center" }}>
          <button
            className="btn btn-primary"
            type="button"
            onClick={executeQuery}
            disabled={executing || query.trim() === ""}
          >
            {executing ? "Running..." : "Run query"}
          </button>
          <button className="btn btn-secondary" type="button" onClick={() => setResult(null)}>
            Clear results
          </button>
          <span style={{ color: "var(--muted)", fontSize: "0.85rem" }}>
            SELECT, WITH, SHOW, and EXPLAIN only. Results cap at 200 rows.
          </span>
        </div>
      </div>

      {result ? (
        <div className="card" style={{ marginBottom: "1rem" }}>
          <div
            style={{
              display: "flex",
              gap: "0.75rem",
              flexWrap: "wrap",
              justifyContent: "space-between",
              marginBottom: "0.75rem",
            }}
          >
            <h2 style={{ margin: 0, fontSize: "1.05rem" }}>Query results</h2>
            <div style={{ color: "var(--muted)", fontSize: "0.85rem" }}>
              {result.row_count} rows{result.truncated ? ` shown of ${result.max_rows}+` : ""} in {result.elapsed_ms} ms
            </div>
          </div>
          {result.columns.length === 0 ? (
            <InlineEmpty message="Query completed with no result columns." />
          ) : result.rows.length === 0 ? (
            <InlineEmpty message="No rows returned." />
          ) : (
            <div className="table-wrap">
              <table className="data">
                <thead>
                  <tr>
                    {result.columns.map((column, index) => (
                      <th key={`${column.name}-${index}`}>{column.name || `column_${index + 1}`}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {result.rows.map((row, rowIndex) => (
                    <tr key={rowIndex}>
                      {row.map((value, columnIndex) => (
                        <td
                          key={columnIndex}
                          style={{
                            maxWidth: "28rem",
                            overflowWrap: "anywhere",
                            whiteSpace: "normal",
                            verticalAlign: "top",
                          }}
                        >
                          <code>{formatCell(value)}</code>
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      ) : null}

      <div className="card">
        <div
          style={{
            display: "flex",
            gap: "0.75rem",
            flexWrap: "wrap",
            justifyContent: "space-between",
            alignItems: "end",
            marginBottom: "0.75rem",
          }}
        >
          <div>
            <h2 style={{ margin: 0, fontSize: "1.05rem" }}>Schema</h2>
            <p style={{ margin: "0.25rem 0 0", color: "var(--muted)" }}>
              {schema.length} tables and views in the public schema
            </p>
          </div>
          <label style={{ display: "grid", gap: "0.35rem", minWidth: "min(100%, 18rem)" }}>
            <span style={{ fontSize: "0.85rem", fontWeight: 600 }}>Filter</span>
            <input
              className="input"
              value={schemaFilter}
              onChange={(event) => setSchemaFilter(event.target.value)}
              placeholder="Table or column"
            />
          </label>
        </div>

        {loadingSchema ? <InlineEmpty message="Loading schema..." /> : null}
        {!loadingSchema && filteredSchema.length === 0 ? (
          <InlineEmpty message="No schema entries match the filter." />
        ) : null}
        {!loadingSchema && filteredSchema.length > 0 ? (
          <div className="table-wrap">
            <table className="data">
              <thead>
                <tr>
                  <th>Table</th>
                  <th>Type</th>
                  <th>Column</th>
                  <th>Data type</th>
                  <th>Nullable</th>
                  <th>Default</th>
                </tr>
              </thead>
              <tbody>
                {filteredSchema.flatMap((table) =>
                  table.columns.map((column, columnIndex) => (
                    <tr key={`${table.name}-${column.name}`}>
                      <td>{columnIndex === 0 ? table.name : ""}</td>
                      <td>{columnIndex === 0 ? table.type : ""}</td>
                      <td>{column.name}</td>
                      <td>{column.data_type}</td>
                      <td>{column.is_nullable ? "Yes" : "No"}</td>
                      <td
                        style={{
                          maxWidth: "26rem",
                          overflowWrap: "anywhere",
                          whiteSpace: "normal",
                        }}
                      >
                        {column.column_default || ""}
                      </td>
                    </tr>
                  )),
                )}
              </tbody>
            </table>
          </div>
        ) : null}
      </div>
    </div>
  );
}
