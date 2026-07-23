import { StrictMode, useEffect, useMemo, useState } from 'react'
import { createRoot } from 'react-dom/client'
import {
  Button, Column, Content, Grid, Header, HeaderName, InlineNotification, Search,
  SideNav, SideNavItems, SideNavLink, SkeletonText, Table, TableBody,
  TableCell, TableContainer, TableHead, TableHeader, TableRow, Tag, Theme, Tile,
} from '@carbon/react'
import { Close, Dashboard, DataVis_1, DocumentExport, DocumentSecurity, Network_3, Renew, Save, Task as TaskIcon } from '@carbon/icons-react'
import '@carbon/styles/css/styles.css'
import './styles.css'

type Stamp = { seconds: number; nanos: number } | null
type Scope = { id: string; name: string; authorization_ref?: string; targets: { kind: number; value: string }[]; created_at: Stamp }
type Task = { id: string; scope_id: string; policy_id: string; state: number; created_at: Stamp; updated_at: Stamp; schedule_id: string }
type Asset = { id: string; scope_id: string; kind: number; value: string; first_seen_at: Stamp; last_seen_at: Stamp }
type Service = { id: string; asset_id: string; transport: string; port: number; service_hint: string; first_seen_at: Stamp; last_seen_at: Stamp }
type Certificate = { id: string; service_id: string; sha256: string; subject: string; issuer: string; dns_names: string[]; not_before: Stamp; not_after: Stamp; first_seen_at: Stamp; last_seen_at: Stamp }
type TechnologyFingerprint = { id: string; name: string; version: string; detector: string; rule_id: string; evidence_id: string }
type Website = { id: string; service_id: string; url: string; status_code: number; title: string; server: string; content_type: string; content_sha256: string; fingerprints: TechnologyFingerprint[]; discovered_paths: string[]; screenshot_evidence_id: string; first_seen_at: Stamp; last_seen_at: Stamp }
type Finding = { id: string; scope_id: string; task_id: string; asset_id: string; observation_id: string; evidence_id: string; detector: string; rule_id: string; title: string; description: string; severity: number; state: number; fingerprint: string; first_seen_at: Stamp; last_seen_at: Stamp }
type Schedule = { id: string; scope_id: string; policy_id: string; interval_seconds: number; enabled: boolean; next_run_at: Stamp; last_task_id: string; created_at: Stamp }
type AssetChange = { id: string; schedule_id: string; task_id: string; asset_id: string; kind: number; detected_at: Stamp }
type ExposureChange = { id: string; schedule_id: string; task_id: string; resource_kind: string; resource_id: string; kind: number; previous_fingerprint: string; current_fingerprint: string; detected_at: Stamp }
type AuditEvent = { id: string; request_id?: string; operation: string; agent_id?: string; skill_name?: string; skill_version?: string; resource_id: string; occurred_at: Stamp }
type Observation = { id: string; task_id: string; asset_id: string; type: string; value: unknown; evidence_id: string; observed_at: Stamp }
type Evidence = { id: string; media_type: string; sha256: string; content_base64: string; created_at: Stamp }
type Overview = {
  counts: { scopes: number; tasks: number; assets: number; services: number; certificates: number; websites: number; findings: number; schedules: number; asset_changes: number; exposure_changes: number; observations: number; evidence: number; notifications_pending: number; notifications_delivered: number; notifications_dead_letter: number }
  scopes: Scope[]
  tasks: Task[]
  assets: Asset[]
  services: Service[]
  certificates: Certificate[]
  websites: Website[]
  findings: Finding[]
  schedules: Schedule[]
  asset_changes: AssetChange[]
  exposure_changes: ExposureChange[]
  audit_events: AuditEvent[]
}

const taskStates: Record<number, { label: string; type: 'gray' | 'blue' | 'green' | 'red' }> = {
  1: { label: 'Queued', type: 'gray' },
  2: { label: 'Running', type: 'blue' },
  3: { label: 'Completed', type: 'green' },
  4: { label: 'Failed', type: 'red' },
  5: { label: 'Canceled', type: 'gray' },
}

type View = 'overview' | 'inventory' | 'risk' | 'tasks' | 'monitoring' | 'audit'
type SavedView = { name: string; view: View; query: string; severity: string; taskState: string }
type Inspector = { kind: 'asset'; asset: Asset } | { kind: 'task'; task: Task } | { kind: 'finding'; finding: Finding } | { kind: 'evidence'; evidence: Evidence }

const navigation: { id: View; label: string; icon: typeof Dashboard }[] = [
  { id: 'overview', label: 'Overview', icon: Dashboard },
  { id: 'inventory', label: 'Inventory', icon: Network_3 },
  { id: 'risk', label: 'Findings', icon: DocumentSecurity },
  { id: 'tasks', label: 'Tasks', icon: TaskIcon },
  { id: 'monitoring', label: 'Monitoring', icon: DataVis_1 },
  { id: 'audit', label: 'Audit', icon: DocumentSecurity },
]

function App() {
  const [data, setData] = useState<Overview | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [query, setQuery] = useState('')
  const [view, setView] = useState<View>(() => parseView(location.hash))
  const [severity, setSeverity] = useState('all')
  const [taskState, setTaskState] = useState('all')
  const [savedViews, setSavedViews] = useState<SavedView[]>(() => readSavedViews())
  const [inspector, setInspector] = useState<Inspector | null>(null)
  const [observations, setObservations] = useState<Observation[]>([])
  const [refresh, setRefresh] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    fetch('/api/v1/overview', { signal: controller.signal })
      .then((response) => {
        if (!response.ok) throw new Error(`Read model returned ${response.status}`)
        return response.json() as Promise<Overview>
      })
      .then((value) => { setData(value); setError(null) })
      .catch((reason: Error) => {
        if (reason.name !== 'AbortError') setError(reason.message)
      })
    return () => controller.abort()
  }, [refresh])

  useEffect(() => {
    const onHash = () => setView(parseView(location.hash))
    window.addEventListener('hashchange', onHash)
    return () => window.removeEventListener('hashchange', onHash)
  }, [])

  useEffect(() => {
    if (!inspector) return
    document.querySelector<HTMLElement>('.inspector')?.focus()
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setInspector(null)
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [inspector])

  const filtered = useMemo(() => {
    const value = query.trim().toLowerCase()
    if (!data) return null
    const contains = (...parts: (string | number | undefined)[]) => !value || parts.some((part) => String(part ?? '').toLowerCase().includes(value))
    return {
      assets: data.assets.filter((item) => contains(item.value, item.id, item.scope_id)),
      services: data.services.filter((item) => contains(item.id, item.asset_id, item.port, item.service_hint)),
      certificates: data.certificates.filter((item) => contains(item.subject, item.issuer, item.sha256, ...item.dns_names)),
      websites: data.websites.filter((item) => contains(item.url, item.title, item.server, ...item.fingerprints.map((fingerprint) => fingerprint.name))),
      findings: data.findings.filter((item) => contains(item.title, item.description, item.rule_id, item.detector) && (severity === 'all' || item.severity === Number(severity))),
      tasks: data.tasks.filter((item) => contains(item.id, item.policy_id, item.scope_id) && (taskState === 'all' || item.state === Number(taskState))),
      audit: data.audit_events.filter((item) => contains(item.operation, item.agent_id, item.skill_name, item.resource_id, item.request_id)),
    }
  }, [data, query, severity, taskState])

  const selectView = (next: View) => {
    location.hash = next
    setView(next)
    setInspector(null)
  }

  const saveView = () => {
    const name = window.prompt('Name this local view')?.trim()
    if (!name) return
    const next = [...savedViews.filter((item) => item.name !== name), { name, view, query, severity, taskState }]
    localStorage.setItem('cyberedge.saved-views', JSON.stringify(next))
    setSavedViews(next)
  }

  const openTask = async (task: Task) => {
    setInspector({ kind: 'task', task })
    setObservations([])
    const response = await fetch(`/api/v1/tasks/${encodeURIComponent(task.id)}/observations`)
    if (!response.ok) {
      setError(`Task observations returned ${response.status}`)
      return
    }
    setError(null)
    setObservations((await response.json() as { observations: Observation[] }).observations)
  }

  const openEvidence = async (evidenceId: string) => {
    const response = await fetch(`/api/v1/evidence/${encodeURIComponent(evidenceId)}`)
    if (!response.ok) {
      setError(`Evidence access returned ${response.status}`)
      return
    }
    setError(null)
    setInspector({ kind: 'evidence', evidence: await response.json() as Evidence })
  }

  return (
    <Theme theme="g100">
      <a className="skip-link" href="#main-content">Skip to main content</a>
      <Header aria-label="CyberEdge read-only observer">
        <HeaderName href="#overview" prefix="">CyberEdge Observer</HeaderName>
        <span className="read-only-mark">READ ONLY</span>
      </Header>
      <SideNav aria-label="Observation navigation" expanded isPersistent>
        <SideNavItems>
          {navigation.map((item) => <SideNavLink key={item.id} href={`#${item.id}`} isActive={view === item.id} renderIcon={item.icon}>{item.label}</SideNavLink>)}
        </SideNavItems>
      </SideNav>
      <nav className="mobile-nav" aria-label="Mobile observation navigation">
        {navigation.map((item) => <a key={item.id} href={`#${item.id}`} aria-current={view === item.id ? 'page' : undefined}>{item.label}</a>)}
      </nav>
      <Content id="main-content">
        <main>
          <section className="page-heading" aria-labelledby="page-title">
            <div><p className="section-label">External attack surface</p><h1 id="page-title">{navigation.find((item) => item.id === view)?.label}</h1><p>{viewDescription(view)}</p></div>
            <div className="heading-actions">
              <Button kind="ghost" size="sm" renderIcon={Renew} onClick={() => setRefresh((value) => value + 1)}>Refresh</Button>
              <Button kind="ghost" size="sm" renderIcon={Save} onClick={saveView}>Save view</Button>
              <Button kind="ghost" size="sm" renderIcon={DocumentExport} disabled={!data} onClick={() => data && exportSnapshot(data)}>Export JSON</Button>
            </div>
          </section>

          <div className="command-bar" aria-label="Global read model controls">
            <Search size="lg" labelText="Search all projections" placeholder="Search assets, findings, tasks, certificates, audit" value={query} onChange={(event) => setQuery(event.target.value)} />
            {view === 'risk' && <label className="compact-filter">Severity<select value={severity} onChange={(event) => setSeverity(event.target.value)}><option value="all">All</option><option value="5">Critical</option><option value="4">High</option><option value="3">Medium</option><option value="2">Low</option><option value="1">Info</option></select></label>}
            {view === 'tasks' && <label className="compact-filter">State<select value={taskState} onChange={(event) => setTaskState(event.target.value)}><option value="all">All</option>{Object.entries(taskStates).map(([id, state]) => <option key={id} value={id}>{state.label}</option>)}</select></label>}
            {savedViews.length > 0 && <label className="compact-filter">Saved view<select defaultValue="" onChange={(event) => { const saved = savedViews.find((item) => item.name === event.target.value); if (saved) { selectView(saved.view); setQuery(saved.query); setSeverity(saved.severity); setTaskState(saved.taskState) } }}><option value="">Choose</option>{savedViews.map((item) => <option key={item.name}>{item.name}</option>)}</select></label>}
          </div>

          {error && <InlineNotification kind="error" title="Read model unavailable" subtitle={error} role="alert" />}
          {!data && !error && <Loading />}
          {data && filtered && <>
            {view === 'overview' && <>
            <Grid condensed className="metric-grid">
              <Metric label="Authorized scopes" value={data.counts.scopes} />
              <Metric label="Observed assets" value={data.counts.assets} />
              <Metric label="Open findings" value={data.findings.filter((finding) => finding.state === 1).length} />
              <Metric label="Evidence objects" value={data.counts.evidence} />
            </Grid>
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
            <section className="data-section"><div className="section-heading"><div><h2>Highest priority findings</h2></div><Button kind="ghost" size="sm" onClick={() => selectView('risk')}>View all</Button></div><FindingTable findings={[...data.findings].sort((a, b) => b.severity - a.severity).slice(0, 8)} assets={data.assets} onSelect={(finding) => setInspector({ kind: 'finding', finding })} /></section>
            </>}

            {view === 'inventory' && <>
              <section className="data-section"><div className="section-heading"><h2>Assets</h2><span>{filtered.assets.length} visible</span></div><AssetTable assets={filtered.assets} onSelect={(asset) => setInspector({ kind: 'asset', asset })} /></section>
              <section className="data-section"><h2>Services</h2><ServiceTable services={filtered.services} assets={data.assets} /></section>
              <section className="data-section"><h2>Certificates</h2><CertificateTable certificates={filtered.certificates} services={data.services} assets={data.assets} /></section>
              <section className="data-section"><h2>Websites</h2><WebsiteTable websites={filtered.websites} /></section>
            </>}
            {view === 'risk' && <section className="data-section"><div className="section-heading"><h2>Evidence-backed findings</h2><span>{filtered.findings.length} visible</span></div><FindingTable findings={filtered.findings} assets={data.assets} onSelect={(finding) => setInspector({ kind: 'finding', finding })} /></section>}
            {view === 'tasks' && <section className="data-section"><div className="section-heading"><h2>Task history</h2><span>{filtered.tasks.length} visible</span></div><TaskTable tasks={filtered.tasks} onSelect={openTask} /></section>}
            {view === 'monitoring' && <section className="data-section"><h2>Monitoring schedules</h2><ScheduleTable schedules={data.schedules} /><div className="subsection-heading"><h3>Asset changes</h3><span>{data.counts.asset_changes} recorded</span></div><ChangeTable changes={data.asset_changes} assets={data.assets} /><div className="subsection-heading"><h3>Exposure changes</h3><span>{data.counts.exposure_changes} recorded</span></div><ExposureChangeTable changes={data.exposure_changes} /><div className="delivery-summary">Notifications: {data.counts.notifications_delivered} delivered, {data.counts.notifications_pending} pending, {data.counts.notifications_dead_letter} dead-lettered</div></section>}
            {view === 'audit' && <section className="data-section"><div className="section-heading"><h2>Invocation audit</h2><span>{filtered.audit.length} visible</span></div><AuditTable events={filtered.audit} /></section>}
          </>}
        </main>
      </Content>
      {inspector && data && <InspectorPanel inspector={inspector} data={data} observations={observations} onClose={() => setInspector(null)} onEvidence={openEvidence} />}
    </Theme>
  )
}

function Metric({ label, value }: { label: string; value: number }) {
  return <Column sm={2} md={2} lg={4}><Tile className="metric"><span>{label}</span><strong>{formatNumber(value)}</strong></Tile></Column>
}

function AssetTable({ assets, onSelect }: { assets: Asset[]; onSelect?: (asset: Asset) => void }) {
  if (assets.length === 0) return <Empty text="No assets match the current read-only view." />
  return <TableContainer><Table size="lg" useZebraStyles={false}>
    <TableHead><TableRow><TableHeader>Asset</TableHeader><TableHeader>Kind</TableHeader><TableHeader>Scope</TableHeader><TableHeader>Last observed</TableHeader></TableRow></TableHead>
    <TableBody>{assets.map((asset) => <TableRow key={asset.id}>
      <TableCell>{onSelect ? <Button kind="ghost" size="sm" onClick={() => onSelect(asset)}>{asset.value}</Button> : <strong>{asset.value}</strong>}<code>{asset.id}</code></TableCell>
      <TableCell>{asset.kind === 1 ? 'Domain' : asset.kind === 2 ? 'IP address' : 'Other'}</TableCell>
      <TableCell><code>{shortId(asset.scope_id)}</code></TableCell>
      <TableCell>{formatTime(asset.last_seen_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function TaskTable({ tasks, onSelect }: { tasks: Task[]; onSelect?: (task: Task) => void }) {
  if (tasks.length === 0) return <Empty text="No discovery tasks have been recorded." />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>Task</TableHeader><TableHeader>Policy</TableHeader><TableHeader>State</TableHeader><TableHeader>Updated</TableHeader></TableRow></TableHead>
    <TableBody>{tasks.map((task) => {
      const state = taskStates[task.state] ?? { label: 'Unknown', type: 'gray' as const }
      return <TableRow key={task.id}>
        <TableCell>{onSelect ? <Button kind="ghost" size="sm" onClick={() => onSelect(task)}>{shortId(task.id)}</Button> : <code>{task.id}</code>}</TableCell><TableCell>{task.policy_id}</TableCell>
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
    <TableHead><TableRow><TableHeader>URL</TableHeader><TableHeader>Status</TableHeader><TableHeader>Title</TableHeader><TableHeader>Technology evidence</TableHeader><TableHeader>Last observed</TableHeader></TableRow></TableHead>
    <TableBody>{websites.map((website) => <TableRow key={website.id}>
      <TableCell><strong>{website.url}</strong><code>{shortId(website.content_sha256)} · {website.discovered_paths.length} crawled paths{website.screenshot_evidence_id ? ' · screenshot retained' : ''}</code></TableCell>
      <TableCell><Tag type={website.status_code < 400 ? 'green' : 'red'}>{website.status_code}</Tag></TableCell>
      <TableCell>{website.title || 'Untitled response'}</TableCell>
      <TableCell>{website.fingerprints.length > 0
        ? website.fingerprints.map((fingerprint) => <Tag key={fingerprint.id} type="blue">{fingerprint.name}{fingerprint.version ? ` ${fingerprint.version}` : ''}</Tag>)
        : website.server || 'Not identified'}</TableCell>
      <TableCell>{formatTime(website.last_seen_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function FindingTable({ findings, assets, onSelect }: { findings: Finding[]; assets: Asset[]; onSelect?: (finding: Finding) => void }) {
  const assetNames = new Map(assets.map((asset) => [asset.id, asset.value]))
  const severityLabels: Record<number, string> = { 1: 'Info', 2: 'Low', 3: 'Medium', 4: 'High', 5: 'Critical' }
  if (findings.length === 0) return <Empty text="No evidence-backed findings have been reported." />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>Finding</TableHeader><TableHeader>Asset</TableHeader><TableHeader>Severity</TableHeader><TableHeader>Detector</TableHeader><TableHeader>Last observed</TableHeader></TableRow></TableHead>
    <TableBody>{findings.map((finding) => <TableRow key={finding.id}>
      <TableCell>{onSelect ? <Button kind="ghost" size="sm" onClick={() => onSelect(finding)}>{finding.title}</Button> : <strong>{finding.title}</strong>}<code>{finding.rule_id}</code></TableCell>
      <TableCell>{assetNames.get(finding.asset_id) ?? finding.asset_id}</TableCell>
      <TableCell><Tag type={finding.severity >= 4 ? 'red' : finding.severity === 3 ? 'magenta' : 'cool-gray'}>{severityLabels[finding.severity] ?? 'Unknown'}</Tag></TableCell>
      <TableCell>{finding.detector}</TableCell>
      <TableCell>{formatTime(finding.last_seen_at)}</TableCell>
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

function ExposureChangeTable({ changes }: { changes: ExposureChange[] }) {
  if (changes.length === 0) return <Empty text="No service or website changes have been detected." />
  const labels: Record<number, string> = { 1: 'Appeared', 2: 'Disappeared', 3: 'Modified' }
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>Resource</TableHeader><TableHeader>Change</TableHeader><TableHeader>Fingerprint</TableHeader><TableHeader>Detected</TableHeader></TableRow></TableHead>
    <TableBody>{changes.map((change) => <TableRow key={change.id}>
      <TableCell><strong>{change.resource_kind}</strong><code>{change.resource_id}</code></TableCell>
      <TableCell><Tag type={change.kind === 2 ? 'red' : change.kind === 3 ? 'blue' : 'green'}>{labels[change.kind] ?? 'Unknown'}</Tag></TableCell>
      <TableCell><code>{shortId(change.current_fingerprint || change.previous_fingerprint)}</code></TableCell>
      <TableCell>{formatTime(change.detected_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function AuditTable({ events }: { events: AuditEvent[] }) {
  if (events.length === 0) return <Empty text="No Agent mutations have been audited." />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>Operation</TableHeader><TableHeader>Agent and Skill</TableHeader><TableHeader>Resource</TableHeader><TableHeader>Occurred</TableHeader></TableRow></TableHead>
    <TableBody>{events.map((event) => <TableRow key={event.id}>
      <TableCell><strong>{event.operation}</strong>{event.request_id && <code>{event.request_id}</code>}</TableCell>
      <TableCell>{event.agent_id ?? 'Restricted'}{event.skill_name && <code>{event.skill_name}@{event.skill_version}</code>}</TableCell>
      <TableCell><code>{event.resource_id}</code></TableCell>
      <TableCell>{formatTime(event.occurred_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function InspectorPanel({ inspector, data, observations, onClose, onEvidence }: { inspector: Inspector; data: Overview; observations: Observation[]; onClose: () => void; onEvidence: (id: string) => void }) {
  let title = 'Details'
  let content: React.ReactNode
  if (inspector.kind === 'asset') {
    const services = data.services.filter((item) => item.asset_id === inspector.asset.id)
    const serviceIds = new Set(services.map((item) => item.id))
    const findings = data.findings.filter((item) => item.asset_id === inspector.asset.id)
    title = inspector.asset.value
    content = <><DefinitionList values={{ 'Asset ID': inspector.asset.id, 'Scope ID': inspector.asset.scope_id, 'Kind': inspector.asset.kind === 1 ? 'Domain' : 'IP address', 'First observed': formatTime(inspector.asset.first_seen_at), 'Last observed': formatTime(inspector.asset.last_seen_at) }} /><h3>Relationships</h3><p>{services.length} services, {data.websites.filter((item) => serviceIds.has(item.service_id)).length} websites, {data.certificates.filter((item) => serviceIds.has(item.service_id)).length} certificates, {findings.length} findings</p><FindingTable findings={findings} assets={data.assets} onSelect={(finding) => onEvidence(finding.evidence_id)} /></>
  } else if (inspector.kind === 'task') {
    title = `Task ${shortId(inspector.task.id)}`
    content = <><DefinitionList values={{ 'Task ID': inspector.task.id, 'Policy': inspector.task.policy_id, 'State': taskStates[inspector.task.state]?.label ?? 'Unknown', 'Scope ID': inspector.task.scope_id, 'Updated': formatTime(inspector.task.updated_at) }} /><h3>Observation timeline</h3>{observations.length === 0 ? <Empty text="No observations are available for this task." /> : <ol className="timeline">{observations.map((item) => <li key={item.id}><time>{formatTime(item.observed_at)}</time><strong>{item.type}</strong><code>{item.asset_id}</code><Button kind="ghost" size="sm" onClick={() => onEvidence(item.evidence_id)}>Open evidence</Button></li>)}</ol>}</>
  } else if (inspector.kind === 'finding') {
    const finding = inspector.finding
    title = finding.title
    content = <><DefinitionList values={{ 'Finding ID': finding.id, 'Rule': finding.rule_id, 'Detector': finding.detector, 'Severity': severityName(finding.severity), 'State': finding.state === 1 ? 'Open' : 'Resolved', 'Asset ID': finding.asset_id, 'Last observed': formatTime(finding.last_seen_at) }} /><p className="inspector-description">{finding.description}</p><Button kind="secondary" onClick={() => onEvidence(finding.evidence_id)}>Inspect supporting evidence</Button></>
  } else {
    title = `Evidence ${shortId(inspector.evidence.id)}`
    content = <><DefinitionList values={{ 'Evidence ID': inspector.evidence.id, 'Media type': inspector.evidence.media_type, 'SHA-256': inspector.evidence.sha256, 'Created': formatTime(inspector.evidence.created_at) }} /><h3>Preview</h3><pre className="evidence-preview">{evidencePreview(inspector.evidence)}</pre></>
  }
  return <aside className="inspector" aria-labelledby="inspector-title" tabIndex={-1}><div className="inspector-header"><h2 id="inspector-title">{title}</h2><Button hasIconOnly kind="ghost" iconDescription="Close details" renderIcon={Close} onClick={onClose} /></div><div className="inspector-body">{content}</div></aside>
}

function DefinitionList({ values }: { values: Record<string, string | number> }) {
  return <dl className="definition-list">{Object.entries(values).map(([label, value]) => <div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl>
}

function Loading() {
  return <div className="loading" aria-label="Loading read model"><SkeletonText heading width="35%" /><SkeletonText paragraph lineCount={6} /></div>
}

function Empty({ text }: { text: string }) {
  return <div className="empty"><strong>Nothing to display</strong><p>{text}</p></div>
}

function parseView(hash: string): View {
  const value = hash.replace('#', '') as View
  return navigation.some((item) => item.id === value) ? value : 'overview'
}

function viewDescription(view: View) {
  const descriptions: Record<View, string> = {
    overview: 'Current exposure posture, authorization coverage, and evidence retention.',
    inventory: 'Connected assets, services, certificates, and websites from authorized scopes.',
    risk: 'Evidence-backed findings with severity and lifecycle state.',
    tasks: 'Deterministic execution history and observation timelines.',
    monitoring: 'Recurring schedules, exposure drift, and notification delivery.',
    audit: 'Agent and Skill provenance for every control-plane mutation.',
  }
  return descriptions[view]
}

function readSavedViews(): SavedView[] {
  try {
    const value = JSON.parse(localStorage.getItem('cyberedge.saved-views') ?? '[]')
    return Array.isArray(value) ? value : []
  } catch {
    return []
  }
}

function exportSnapshot(data: Overview) {
  const blob = new Blob([JSON.stringify({ generated_at: new Date().toISOString(), read_only: true, ...data }, null, 2)], { type: 'application/json' })
  const link = document.createElement('a')
  link.href = URL.createObjectURL(blob)
  link.download = `cyberedge-observer-${new Date().toISOString().slice(0, 10)}.json`
  link.click()
  URL.revokeObjectURL(link.href)
}

function evidencePreview(evidence: Evidence) {
  if (!evidence.media_type.startsWith('text/') && !evidence.media_type.includes('json')) return 'Binary evidence preview is unavailable. Verify it using the retained digest.'
  try {
    const bytes = Uint8Array.from(atob(evidence.content_base64), (value) => value.charCodeAt(0))
    const value = new TextDecoder().decode(bytes.slice(0, 32768))
    return bytes.length > 32768 ? `${value}\n[Preview truncated at 32 KiB]` : value
  } catch {
    return 'Evidence content could not be decoded.'
  }
}

function severityName(value: number) { return ({ 1: 'Info', 2: 'Low', 3: 'Medium', 4: 'High', 5: 'Critical' } as Record<number, string>)[value] ?? 'Unknown' }

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
