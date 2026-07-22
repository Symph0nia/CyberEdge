import { StrictMode, useEffect, useMemo, useState } from 'react'
import { createRoot } from 'react-dom/client'
import {
  Column, Content, Grid, Header, HeaderName, InlineNotification, Search,
  SideNav, SideNavItems, SideNavLink, SkeletonText, Table, TableBody,
  TableCell, TableContainer, TableHead, TableHeader, TableRow, Tag, Theme, Tile,
} from '@carbon/react'
import { Dashboard, DataVis_1, DocumentSecurity, Network_3, Task as TaskIcon } from '@carbon/icons-react'
import '@carbon/styles/css/styles.css'
import './styles.css'

type Stamp = { seconds: number; nanos: number } | null
type Scope = { id: string; name: string; authorization_ref: string; targets: { kind: number; value: string }[]; created_at: Stamp }
type Task = { id: string; scope_id: string; policy_id: string; state: number; created_at: Stamp; updated_at: Stamp }
type Asset = { id: string; scope_id: string; kind: number; value: string; first_seen_at: Stamp; last_seen_at: Stamp }
type AuditEvent = { id: string; request_id: string; operation: string; agent_id: string; skill_name: string; skill_version: string; resource_id: string; occurred_at: Stamp }
type Overview = {
  counts: { scopes: number; tasks: number; assets: number; observations: number; evidence: number }
  scopes: Scope[]
  tasks: Task[]
  assets: Asset[]
  audit_events: AuditEvent[]
}

const taskStates: Record<number, { label: string; type: 'gray' | 'blue' | 'green' | 'red' }> = {
  1: { label: 'Queued', type: 'gray' },
  2: { label: 'Running', type: 'blue' },
  3: { label: 'Completed', type: 'green' },
  4: { label: 'Failed', type: 'red' },
  5: { label: 'Canceled', type: 'gray' },
}

function App() {
  const [data, setData] = useState<Overview | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [query, setQuery] = useState('')

  useEffect(() => {
    const controller = new AbortController()
    fetch('/api/v1/overview', { signal: controller.signal })
      .then((response) => {
        if (!response.ok) throw new Error(`Read model returned ${response.status}`)
        return response.json() as Promise<Overview>
      })
      .then(setData)
      .catch((reason: Error) => {
        if (reason.name !== 'AbortError') setError(reason.message)
      })
    return () => controller.abort()
  }, [])

  const assets = useMemo(() => {
    const value = query.trim().toLowerCase()
    return data?.assets.filter((asset) => asset.value.toLowerCase().includes(value)) ?? []
  }, [data, query])

  return (
    <Theme theme="g100">
      <Header aria-label="CyberEdge read-only observer">
        <HeaderName href="#overview" prefix="">CyberEdge</HeaderName>
        <span className="read-only-mark">READ ONLY</span>
      </Header>
      <SideNav aria-label="Observation navigation" expanded isPersistent>
        <SideNavItems>
          <SideNavLink href="#overview" renderIcon={Dashboard}>Overview</SideNavLink>
          <SideNavLink href="#assets" renderIcon={Network_3}>Assets</SideNavLink>
          <SideNavLink href="#tasks" renderIcon={TaskIcon}>Tasks</SideNavLink>
          <SideNavLink href="#evidence" renderIcon={DocumentSecurity}>Evidence</SideNavLink>
          <SideNavLink href="#audit" renderIcon={DocumentSecurity}>Audit</SideNavLink>
          <SideNavLink href="#coverage" renderIcon={DataVis_1}>Coverage</SideNavLink>
        </SideNavItems>
      </SideNav>
      <Content id="main-content">
        <main>
          <section id="overview" className="page-heading" aria-labelledby="overview-title">
            <p className="section-label">External attack surface</p>
            <h1 id="overview-title">Operational overview</h1>
            <p>Live inventory derived from authorized scopes and immutable discovery evidence.</p>
          </section>

          {error && <InlineNotification kind="error" title="Read model unavailable" subtitle={error} role="alert" />}
          {!data && !error && <Loading />}
          {data && <>
            <Grid condensed className="metric-grid">
              <Metric label="Authorized scopes" value={data.counts.scopes} />
              <Metric label="Observed assets" value={data.counts.assets} />
              <Metric label="Observations" value={data.counts.observations} />
              <Metric label="Evidence objects" value={data.counts.evidence} />
            </Grid>

            <section id="assets" className="data-section" aria-labelledby="assets-title">
              <div className="section-heading">
                <div><p className="section-label">Inventory</p><h2 id="assets-title">Latest assets</h2></div>
                <Search size="sm" labelText="Filter assets" placeholder="Filter domain or IP" value={query} onChange={(event) => setQuery(event.target.value)} />
              </div>
              <AssetTable assets={assets} />
            </section>

            <section id="tasks" className="data-section" aria-labelledby="tasks-title">
              <div className="section-heading"><div><p className="section-label">Execution</p><h2 id="tasks-title">Task history</h2></div></div>
              <TaskTable tasks={data.tasks} />
            </section>

            <section id="audit" className="data-section" aria-labelledby="audit-title">
              <div className="section-heading"><div><p className="section-label">Agent control plane</p><h2 id="audit-title">Invocation audit</h2></div></div>
              <AuditTable events={data.audit_events} />
            </section>

            <Grid condensed className="lower-grid">
              <Column sm={4} md={4} lg={8}>
                <section id="coverage" className="data-section" aria-labelledby="scope-title">
                  <p className="section-label">Authorization boundary</p>
                  <h2 id="scope-title">Scopes</h2>
                  <div className="scope-list">
                    {data.scopes.length === 0 ? <Empty text="No authorized scopes have been recorded." /> : data.scopes.map((scope) => (
                      <article key={scope.id} className="scope-row">
                        <div><strong>{scope.name}</strong><code>{scope.id}</code></div>
                        <div className="target-list">{scope.targets.map((target) => <Tag key={`${target.kind}:${target.value}`} type="cool-gray">{target.value}</Tag>)}</div>
                      </article>
                    ))}
                  </div>
                </section>
              </Column>
              <Column sm={4} md={4} lg={8}>
                <section id="evidence" className="data-section evidence-summary" aria-labelledby="evidence-title">
                  <p className="section-label">Chain of custody</p>
                  <h2 id="evidence-title">Evidence retention</h2>
                  <p className="evidence-number">{formatNumber(data.counts.evidence)}</p>
                  <p>Content-addressed JSON objects are linked from every observation. Mutations are unavailable on this surface.</p>
                </section>
              </Column>
            </Grid>
          </>}
        </main>
      </Content>
    </Theme>
  )
}

function Metric({ label, value }: { label: string; value: number }) {
  return <Column sm={2} md={2} lg={4}><Tile className="metric"><span>{label}</span><strong>{formatNumber(value)}</strong></Tile></Column>
}

function AssetTable({ assets }: { assets: Asset[] }) {
  if (assets.length === 0) return <Empty text="No assets match the current read-only view." />
  return <TableContainer><Table size="lg" useZebraStyles={false}>
    <TableHead><TableRow><TableHeader>Asset</TableHeader><TableHeader>Kind</TableHeader><TableHeader>Scope</TableHeader><TableHeader>Last observed</TableHeader></TableRow></TableHead>
    <TableBody>{assets.map((asset) => <TableRow key={asset.id}>
      <TableCell><strong>{asset.value}</strong><code>{asset.id}</code></TableCell>
      <TableCell>{asset.kind === 1 ? 'Domain' : asset.kind === 2 ? 'IP address' : 'Other'}</TableCell>
      <TableCell><code>{shortId(asset.scope_id)}</code></TableCell>
      <TableCell>{formatTime(asset.last_seen_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function TaskTable({ tasks }: { tasks: Task[] }) {
  if (tasks.length === 0) return <Empty text="No discovery tasks have been recorded." />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>Task</TableHeader><TableHeader>Policy</TableHeader><TableHeader>State</TableHeader><TableHeader>Updated</TableHeader></TableRow></TableHead>
    <TableBody>{tasks.map((task) => {
      const state = taskStates[task.state] ?? { label: 'Unknown', type: 'gray' as const }
      return <TableRow key={task.id}>
        <TableCell><code>{task.id}</code></TableCell><TableCell>{task.policy_id}</TableCell>
        <TableCell><Tag type={state.type}>{state.label}</Tag></TableCell><TableCell>{formatTime(task.updated_at)}</TableCell>
      </TableRow>
    })}</TableBody>
  </Table></TableContainer>
}

function AuditTable({ events }: { events: AuditEvent[] }) {
  if (events.length === 0) return <Empty text="No Agent mutations have been audited." />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>Operation</TableHeader><TableHeader>Agent and Skill</TableHeader><TableHeader>Resource</TableHeader><TableHeader>Occurred</TableHeader></TableRow></TableHead>
    <TableBody>{events.map((event) => <TableRow key={event.id}>
      <TableCell><strong>{event.operation}</strong><code>{event.request_id}</code></TableCell>
      <TableCell>{event.agent_id}<code>{event.skill_name}@{event.skill_version}</code></TableCell>
      <TableCell><code>{event.resource_id}</code></TableCell>
      <TableCell>{formatTime(event.occurred_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function Loading() {
  return <div className="loading" aria-label="Loading read model"><SkeletonText heading width="35%" /><SkeletonText paragraph lineCount={6} /></div>
}

function Empty({ text }: { text: string }) {
  return <div className="empty"><strong>Nothing to display</strong><p>{text}</p></div>
}

function shortId(value: string) { return value.length > 22 ? `${value.slice(0, 18)}...` : value }
function formatNumber(value: number) { return new Intl.NumberFormat('en-US').format(value) }
function formatTime(stamp: Stamp) {
  if (!stamp) return 'Unknown'
  return new Intl.DateTimeFormat('en-GB', { dateStyle: 'medium', timeStyle: 'medium' }).format(new Date(stamp.seconds * 1000))
}

createRoot(document.getElementById('root')!).render(<StrictMode><App /></StrictMode>)
