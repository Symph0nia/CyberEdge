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
type Task = { id: string; scope_id: string; policy_id: string; state: number; created_at: Stamp; updated_at: Stamp; schedule_id: string }
type Asset = { id: string; scope_id: string; kind: number; value: string; first_seen_at: Stamp; last_seen_at: Stamp }
type Service = { id: string; asset_id: string; transport: string; port: number; service_hint: string; first_seen_at: Stamp; last_seen_at: Stamp }
type Certificate = { id: string; service_id: string; sha256: string; subject: string; issuer: string; dns_names: string[]; not_before: Stamp; not_after: Stamp; first_seen_at: Stamp; last_seen_at: Stamp }
type Website = { id: string; service_id: string; url: string; status_code: number; title: string; server: string; content_type: string; content_sha256: string; first_seen_at: Stamp; last_seen_at: Stamp }
type Schedule = { id: string; scope_id: string; policy_id: string; interval_seconds: number; enabled: boolean; next_run_at: Stamp; last_task_id: string; created_at: Stamp }
type AssetChange = { id: string; schedule_id: string; task_id: string; asset_id: string; kind: number; detected_at: Stamp }
type AuditEvent = { id: string; request_id: string; operation: string; agent_id: string; skill_name: string; skill_version: string; resource_id: string; occurred_at: Stamp }
type Overview = {
  counts: { scopes: number; tasks: number; assets: number; services: number; certificates: number; websites: number; schedules: number; asset_changes: number; observations: number; evidence: number }
  scopes: Scope[]
  tasks: Task[]
  assets: Asset[]
  services: Service[]
  certificates: Certificate[]
  websites: Website[]
  schedules: Schedule[]
  asset_changes: AssetChange[]
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
          <SideNavLink href="#services" renderIcon={Network_3}>Services</SideNavLink>
          <SideNavLink href="#certificates" renderIcon={DocumentSecurity}>Certificates</SideNavLink>
          <SideNavLink href="#websites" renderIcon={Network_3}>Websites</SideNavLink>
          <SideNavLink href="#tasks" renderIcon={TaskIcon}>Tasks</SideNavLink>
          <SideNavLink href="#monitoring" renderIcon={DataVis_1}>Monitoring</SideNavLink>
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
              <Metric label="Open services" value={data.counts.services} />
              <Metric label="TLS certificates" value={data.counts.certificates} />
            </Grid>

            <section id="assets" className="data-section" aria-labelledby="assets-title">
              <div className="section-heading">
                <div><p className="section-label">Inventory</p><h2 id="assets-title">Latest assets</h2></div>
                <Search size="sm" labelText="Filter assets" placeholder="Filter domain or IP" value={query} onChange={(event) => setQuery(event.target.value)} />
              </div>
              <AssetTable assets={assets} />
            </section>

            <section id="services" className="data-section" aria-labelledby="services-title">
              <div className="section-heading"><div><p className="section-label">Exposure</p><h2 id="services-title">Observed services</h2></div></div>
              <ServiceTable services={data.services} assets={data.assets} />
            </section>

            <section id="certificates" className="data-section" aria-labelledby="certificates-title">
              <div className="section-heading"><div><p className="section-label">Transport security</p><h2 id="certificates-title">Observed TLS certificates</h2></div></div>
              <CertificateTable certificates={data.certificates} services={data.services} assets={data.assets} />
            </section>

            <section id="websites" className="data-section" aria-labelledby="websites-title">
              <div className="section-heading"><div><p className="section-label">HTTP observation</p><h2 id="websites-title">Observed websites</h2></div></div>
              <WebsiteTable websites={data.websites} />
            </section>

            <section id="tasks" className="data-section" aria-labelledby="tasks-title">
              <div className="section-heading"><div><p className="section-label">Execution</p><h2 id="tasks-title">Task history</h2></div></div>
              <TaskTable tasks={data.tasks} />
            </section>

            <section id="monitoring" className="data-section" aria-labelledby="monitoring-title">
              <div className="section-heading"><div><p className="section-label">Recurring observation</p><h2 id="monitoring-title">Monitoring schedules</h2></div></div>
              <ScheduleTable schedules={data.schedules} />
              <div className="subsection-heading"><p className="section-label">Detected drift</p><h3>Asset changes</h3><span>{formatNumber(data.counts.asset_changes)} recorded</span></div>
              <ChangeTable changes={data.asset_changes} assets={data.assets} />
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
                  <p>Content-addressed JSON and binary evidence objects are linked from observations. Mutations are unavailable on this surface.</p>
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

function ServiceTable({ services, assets }: { services: Service[]; assets: Asset[] }) {
  const assetNames = new Map(assets.map((asset) => [asset.id, asset.value]))
  if (services.length === 0) return <Empty text="No open baseline TCP services have been observed." />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>Asset</TableHeader><TableHeader>Endpoint</TableHeader><TableHeader>Service hint</TableHeader><TableHeader>Last observed</TableHeader></TableRow></TableHead>
    <TableBody>{services.map((service) => <TableRow key={service.id}>
      <TableCell><strong>{assetNames.get(service.asset_id) ?? 'Retained asset'}</strong><code>{service.asset_id}</code></TableCell>
      <TableCell><code>{service.transport}/{service.port}</code></TableCell>
      <TableCell>{service.service_hint}</TableCell>
      <TableCell>{formatTime(service.last_seen_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function CertificateTable({ certificates, services, assets }: { certificates: Certificate[]; services: Service[]; assets: Asset[] }) {
  const assetNames = new Map(assets.map((asset) => [asset.id, asset.value]))
  const endpoints = new Map(services.map((service) => [service.id, `${assetNames.get(service.asset_id) ?? service.asset_id}:${service.port}`]))
  if (certificates.length === 0) return <Empty text="No TLS certificates have been observed." />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>Endpoint</TableHeader><TableHeader>Subject</TableHeader><TableHeader>Issuer</TableHeader><TableHeader>Expires</TableHeader></TableRow></TableHead>
    <TableBody>{certificates.map((certificate) => <TableRow key={certificate.id}>
      <TableCell><strong>{endpoints.get(certificate.service_id) ?? 'Retained service'}</strong><code>{shortId(certificate.sha256)}</code></TableCell>
      <TableCell>{certificate.subject}<code>{certificate.dns_names.join(', ') || 'No DNS SAN'}</code></TableCell>
      <TableCell>{certificate.issuer}</TableCell>
      <TableCell>{formatTime(certificate.not_after)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function WebsiteTable({ websites }: { websites: Website[] }) {
  if (websites.length === 0) return <Empty text="No HTTP responses have been observed." />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>URL</TableHeader><TableHeader>Status</TableHeader><TableHeader>Title</TableHeader><TableHeader>Server hint</TableHeader><TableHeader>Last observed</TableHeader></TableRow></TableHead>
    <TableBody>{websites.map((website) => <TableRow key={website.id}>
      <TableCell><strong>{website.url}</strong><code>{shortId(website.content_sha256)}</code></TableCell>
      <TableCell><Tag type={website.status_code < 400 ? 'green' : 'red'}>{website.status_code}</Tag></TableCell>
      <TableCell>{website.title || 'Untitled response'}</TableCell>
      <TableCell>{website.server || 'Not disclosed'}</TableCell>
      <TableCell>{formatTime(website.last_seen_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function ScheduleTable({ schedules }: { schedules: Schedule[] }) {
  if (schedules.length === 0) return <Empty text="No recurring discovery schedules have been recorded." />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>Schedule</TableHeader><TableHeader>Policy</TableHeader><TableHeader>Cadence</TableHeader><TableHeader>Next run</TableHeader><TableHeader>Last task</TableHeader></TableRow></TableHead>
    <TableBody>{schedules.map((schedule) => <TableRow key={schedule.id}>
      <TableCell><code>{schedule.id}</code></TableCell>
      <TableCell>{schedule.policy_id}</TableCell>
      <TableCell><Tag type={schedule.enabled ? 'green' : 'gray'}>{schedule.enabled ? `Every ${formatInterval(schedule.interval_seconds)}` : 'Disabled'}</Tag></TableCell>
      <TableCell>{formatTime(schedule.next_run_at)}</TableCell>
      <TableCell><code>{schedule.last_task_id ? shortId(schedule.last_task_id) : 'Not run'}</code></TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function ChangeTable({ changes, assets }: { changes: AssetChange[]; assets: Asset[] }) {
  const assetNames = new Map(assets.map((asset) => [asset.id, asset.value]))
  if (changes.length === 0) return <Empty text="No drift has been detected after a completed monitoring baseline." />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>Change</TableHeader><TableHeader>Asset</TableHeader><TableHeader>Schedule</TableHeader><TableHeader>Detected</TableHeader></TableRow></TableHead>
    <TableBody>{changes.map((change) => <TableRow key={change.id}>
      <TableCell><Tag type={change.kind === 1 ? 'green' : 'red'}>{change.kind === 1 ? 'Appeared' : 'Disappeared'}</Tag></TableCell>
      <TableCell><strong>{assetNames.get(change.asset_id) ?? 'Retained asset'}</strong><code>{change.asset_id}</code></TableCell>
      <TableCell><code>{shortId(change.schedule_id)}</code></TableCell>
      <TableCell>{formatTime(change.detected_at)}</TableCell>
    </TableRow>)}</TableBody>
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
function formatInterval(seconds: number) {
  if (seconds % 86400 === 0) return `${seconds / 86400}d`
  if (seconds % 3600 === 0) return `${seconds / 3600}h`
  if (seconds % 60 === 0) return `${seconds / 60}m`
  return `${seconds}s`
}
function formatTime(stamp: Stamp) {
  if (!stamp) return 'Unknown'
  return new Intl.DateTimeFormat('en-GB', { dateStyle: 'medium', timeStyle: 'medium' }).format(new Date(stamp.seconds * 1000))
}

createRoot(document.getElementById('root')!).render(<StrictMode><App /></StrictMode>)
