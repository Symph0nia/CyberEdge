import { createContext, StrictMode, useContext, useEffect, useMemo, useState } from 'react'
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

type View = 'overview' | 'inventory' | 'risk' | 'tasks' | 'monitoring' | 'audit'
type Locale = 'en' | 'zh-CN'
type Translate = (key: string, values?: Record<string, string | number>) => string
type SavedView = { name: string; view: View; query: string; severity: string; taskState: string }
type Inspector = { kind: 'asset'; asset: Asset } | { kind: 'task'; task: Task } | { kind: 'finding'; finding: Finding } | { kind: 'evidence'; evidence: Evidence }

const navigation: { id: View; icon: typeof Dashboard }[] = [
  { id: 'overview', icon: Dashboard },
  { id: 'inventory', icon: Network_3 },
  { id: 'risk', icon: DocumentSecurity },
  { id: 'tasks', icon: TaskIcon },
  { id: 'monitoring', icon: DataVis_1 },
  { id: 'audit', icon: DocumentSecurity },
]

function App() {
  const { locale, setLocale, t } = useI18n()
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
        if (!response.ok) throw new Error(t('Read model returned {status}', { status: response.status }))
        return response.json() as Promise<Overview>
      })
      .then((value) => { setData(value); setError(null) })
      .catch((reason: Error) => {
        if (reason.name !== 'AbortError') setError(reason.message)
      })
    return () => controller.abort()
  }, [refresh, t])

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
    const name = window.prompt(t('Name this local view'))?.trim()
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
      setError(t('Task observations returned {status}', { status: response.status }))
      return
    }
    setError(null)
    setObservations((await response.json() as { observations: Observation[] }).observations)
  }

  const openEvidence = async (evidenceId: string) => {
    const response = await fetch(`/api/v1/evidence/${encodeURIComponent(evidenceId)}`)
    if (!response.ok) {
      setError(t('Evidence access returned {status}', { status: response.status }))
      return
    }
    setError(null)
    setInspector({ kind: 'evidence', evidence: await response.json() as Evidence })
  }

  return (
    <Theme theme="g100">
      <a className="skip-link" href="#main-content">{t('Skip to main content')}</a>
      <Header aria-label={t('CyberEdge read-only observer')}>
        <HeaderName href="#overview" prefix="">{t('CyberEdge Observer')}</HeaderName>
        <span className="read-only-mark">{t('READ ONLY')}</span>
        <Button className="locale-switch" kind="ghost" size="sm" onClick={() => setLocale(locale === 'en' ? 'zh-CN' : 'en')} aria-label={t('Switch language')}>
          {locale === 'en' ? '中文' : 'EN'}
        </Button>
      </Header>
      <SideNav aria-label={t('Observation navigation')} expanded isPersistent>
        <SideNavItems>
          {navigation.map((item) => <SideNavLink key={item.id} href={`#${item.id}`} isActive={view === item.id} renderIcon={item.icon}>{viewLabel(item.id, t)}</SideNavLink>)}
        </SideNavItems>
      </SideNav>
      <nav className="mobile-nav" aria-label={t('Mobile observation navigation')}>
        {navigation.map((item) => <a key={item.id} href={`#${item.id}`} aria-current={view === item.id ? 'page' : undefined}>{viewLabel(item.id, t)}</a>)}
      </nav>
      <Content id="main-content">
        <main>
          <section className="page-heading" aria-labelledby="page-title">
            <div><p className="section-label">{t('External attack surface')}</p><h1 id="page-title">{viewLabel(view, t)}</h1><p>{viewDescription(view, t)}</p></div>
            <div className="heading-actions">
              <Button kind="ghost" size="sm" renderIcon={Renew} onClick={() => setRefresh((value) => value + 1)}>{t('Refresh')}</Button>
              <Button kind="ghost" size="sm" renderIcon={Save} onClick={saveView}>{t('Save view')}</Button>
              <Button kind="ghost" size="sm" renderIcon={DocumentExport} disabled={!data} onClick={() => data && exportSnapshot(data)}>{t('Export JSON')}</Button>
            </div>
          </section>

          <div className="command-bar" aria-label="Global read model controls">
            <Search size="lg" labelText={t('Search all projections')} placeholder={t('Search assets, findings, tasks, certificates, audit')} value={query} onChange={(event) => setQuery(event.target.value)} />
            {view === 'risk' && <label className="compact-filter">{t('Severity')}<select value={severity} onChange={(event) => setSeverity(event.target.value)}><option value="all">{t('All')}</option><option value="5">{t('Critical')}</option><option value="4">{t('High')}</option><option value="3">{t('Medium')}</option><option value="2">{t('Low')}</option><option value="1">{t('Info')}</option></select></label>}
            {view === 'tasks' && <label className="compact-filter">{t('State')}<select value={taskState} onChange={(event) => setTaskState(event.target.value)}><option value="all">{t('All')}</option>{[1, 2, 3, 4, 5].map((id) => <option key={id} value={id}>{taskStateMeta(id, t).label}</option>)}</select></label>}
            {savedViews.length > 0 && <label className="compact-filter">{t('Saved view')}<select defaultValue="" onChange={(event) => { const saved = savedViews.find((item) => item.name === event.target.value); if (saved) { selectView(saved.view); setQuery(saved.query); setSeverity(saved.severity); setTaskState(saved.taskState) } }}><option value="">{t('Choose')}</option>{savedViews.map((item) => <option key={item.name}>{item.name}</option>)}</select></label>}
          </div>

          {error && <InlineNotification kind="error" title={t('Read model unavailable')} subtitle={error} role="alert" />}
          {!data && !error && <Loading />}
          {data && filtered && <>
            {view === 'overview' && <>
            <Grid condensed className="metric-grid">
              <Metric label={t('Authorized scopes')} value={data.counts.scopes} />
              <Metric label={t('Observed assets')} value={data.counts.assets} />
              <Metric label={t('Open findings')} value={data.findings.filter((finding) => finding.state === 1).length} />
              <Metric label={t('Evidence objects')} value={data.counts.evidence} />
            </Grid>
            <Grid condensed className="lower-grid">
              <Column sm={4} md={4} lg={8}>
                <section id="coverage" className="data-section" aria-labelledby="scope-title">
                  <p className="section-label">{t('Authorization boundary')}</p>
                  <h2 id="scope-title">{t('Scopes')}</h2>
                  <div className="scope-list">
                    {data.scopes.length === 0 ? <Empty text={t('No authorized scopes have been recorded.')} /> : data.scopes.map((scope) => (
                      <article key={scope.id} className="scope-row">
                        <div><strong>{t(scope.name)}</strong><code>{scope.id}</code></div>
                        <div className="target-list">{scope.targets.map((target) => <Tag key={`${target.kind}:${target.value}`} type="cool-gray">{target.value}</Tag>)}</div>
                      </article>
                    ))}
                  </div>
                </section>
              </Column>
              <Column sm={4} md={4} lg={8}>
                <section id="evidence" className="data-section evidence-summary" aria-labelledby="evidence-title">
                  <p className="section-label">{t('Chain of custody')}</p>
                  <h2 id="evidence-title">{t('Evidence retention')}</h2>
                  <p className="evidence-number">{formatNumber(data.counts.evidence)}</p>
                  <p>{t('Content-addressed JSON and binary evidence objects are linked from observations. Mutations are unavailable on this surface.')}</p>
                </section>
              </Column>
            </Grid>
            <section className="data-section"><div className="section-heading"><div><h2>{t('Highest priority findings')}</h2></div><Button kind="ghost" size="sm" onClick={() => selectView('risk')}>{t('View all')}</Button></div><FindingTable findings={[...data.findings].sort((a, b) => b.severity - a.severity).slice(0, 8)} assets={data.assets} onSelect={(finding) => setInspector({ kind: 'finding', finding })} /></section>
            </>}

            {view === 'inventory' && <>
              <section className="data-section"><div className="section-heading"><h2>{t('Assets')}</h2><span>{t('{count} visible', { count: filtered.assets.length })}</span></div><AssetTable assets={filtered.assets} onSelect={(asset) => setInspector({ kind: 'asset', asset })} /></section>
              <section className="data-section"><h2>{t('Services')}</h2><ServiceTable services={filtered.services} assets={data.assets} /></section>
              <section className="data-section"><h2>{t('Certificates')}</h2><CertificateTable certificates={filtered.certificates} services={data.services} assets={data.assets} /></section>
              <section className="data-section"><h2>{t('Websites')}</h2><WebsiteTable websites={filtered.websites} /></section>
            </>}
            {view === 'risk' && <section className="data-section"><div className="section-heading"><h2>{t('Evidence-backed findings')}</h2><span>{t('{count} visible', { count: filtered.findings.length })}</span></div><FindingTable findings={filtered.findings} assets={data.assets} onSelect={(finding) => setInspector({ kind: 'finding', finding })} /></section>}
            {view === 'tasks' && <section className="data-section"><div className="section-heading"><h2>{t('Task history')}</h2><span>{t('{count} visible', { count: filtered.tasks.length })}</span></div><TaskTable tasks={filtered.tasks} onSelect={openTask} /></section>}
            {view === 'monitoring' && <section className="data-section"><h2>{t('Monitoring schedules')}</h2><ScheduleTable schedules={data.schedules} /><div className="subsection-heading"><h3>{t('Asset changes')}</h3><span>{t('{count} recorded', { count: data.counts.asset_changes })}</span></div><ChangeTable changes={data.asset_changes} assets={data.assets} /><div className="subsection-heading"><h3>{t('Exposure changes')}</h3><span>{t('{count} recorded', { count: data.counts.exposure_changes })}</span></div><ExposureChangeTable changes={data.exposure_changes} /><div className="delivery-summary">{t('Notifications: {delivered} delivered, {pending} pending, {dead} dead-lettered', { delivered: data.counts.notifications_delivered, pending: data.counts.notifications_pending, dead: data.counts.notifications_dead_letter })}</div></section>}
            {view === 'audit' && <section className="data-section"><div className="section-heading"><h2>{t('Invocation audit')}</h2><span>{t('{count} visible', { count: filtered.audit.length })}</span></div><AuditTable events={filtered.audit} /></section>}
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
  const { t } = useI18n()
  if (assets.length === 0) return <Empty text={t('No assets match the current read-only view.')} />
  return <TableContainer><Table size="lg" useZebraStyles={false}>
    <TableHead><TableRow><TableHeader>{t('Asset')}</TableHeader><TableHeader>{t('Kind')}</TableHeader><TableHeader>{t('Scope')}</TableHeader><TableHeader>{t('Last observed')}</TableHeader></TableRow></TableHead>
    <TableBody>{assets.map((asset) => <TableRow key={asset.id}>
      <TableCell>{onSelect ? <Button kind="ghost" size="sm" onClick={() => onSelect(asset)}>{asset.value}</Button> : <strong>{asset.value}</strong>}<code>{asset.id}</code></TableCell>
      <TableCell>{asset.kind === 1 ? t('Domain') : asset.kind === 2 ? t('IP address') : t('Other')}</TableCell>
      <TableCell><code>{shortId(asset.scope_id)}</code></TableCell>
      <TableCell>{formatTime(asset.last_seen_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function TaskTable({ tasks, onSelect }: { tasks: Task[]; onSelect?: (task: Task) => void }) {
  const { t } = useI18n()
  if (tasks.length === 0) return <Empty text={t('No discovery tasks have been recorded.')} />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>{t('Task')}</TableHeader><TableHeader>{t('Policy')}</TableHeader><TableHeader>{t('State')}</TableHeader><TableHeader>{t('Updated')}</TableHeader></TableRow></TableHead>
    <TableBody>{tasks.map((task) => {
      const state = taskStateMeta(task.state, t)
      return <TableRow key={task.id}>
        <TableCell>{onSelect ? <Button kind="ghost" size="sm" onClick={() => onSelect(task)}>{shortId(task.id)}</Button> : <code>{task.id}</code>}</TableCell><TableCell>{t(task.policy_id)}</TableCell>
        <TableCell><Tag type={state.type}>{state.label}</Tag></TableCell><TableCell>{formatTime(task.updated_at)}</TableCell>
      </TableRow>
    })}</TableBody>
  </Table></TableContainer>
}

function ServiceTable({ services, assets }: { services: Service[]; assets: Asset[] }) {
  const { t } = useI18n()
  const assetNames = new Map(assets.map((asset) => [asset.id, asset.value]))
  if (services.length === 0) return <Empty text={t('No open baseline TCP services have been observed.')} />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>{t('Asset')}</TableHeader><TableHeader>{t('Endpoint')}</TableHeader><TableHeader>{t('Service hint')}</TableHeader><TableHeader>{t('Last observed')}</TableHeader></TableRow></TableHead>
    <TableBody>{services.map((service) => <TableRow key={service.id}>
      <TableCell><strong>{assetNames.get(service.asset_id) ?? t('Retained asset')}</strong><code>{service.asset_id}</code></TableCell>
      <TableCell><code>{service.transport}/{service.port}</code></TableCell>
      <TableCell>{service.service_hint}</TableCell>
      <TableCell>{formatTime(service.last_seen_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function CertificateTable({ certificates, services, assets }: { certificates: Certificate[]; services: Service[]; assets: Asset[] }) {
  const { t } = useI18n()
  const assetNames = new Map(assets.map((asset) => [asset.id, asset.value]))
  const endpoints = new Map(services.map((service) => [service.id, `${assetNames.get(service.asset_id) ?? service.asset_id}:${service.port}`]))
  if (certificates.length === 0) return <Empty text={t('No TLS certificates have been observed.')} />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>{t('Endpoint')}</TableHeader><TableHeader>{t('Subject')}</TableHeader><TableHeader>{t('Issuer')}</TableHeader><TableHeader>{t('Expires')}</TableHeader></TableRow></TableHead>
    <TableBody>{certificates.map((certificate) => <TableRow key={certificate.id}>
      <TableCell><strong>{endpoints.get(certificate.service_id) ?? t('Retained service')}</strong><code>{shortId(certificate.sha256)}</code></TableCell>
      <TableCell>{certificate.subject}<code>{certificate.dns_names.join(', ') || t('No DNS SAN')}</code></TableCell>
      <TableCell>{certificate.issuer}</TableCell>
      <TableCell>{formatTime(certificate.not_after)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function WebsiteTable({ websites }: { websites: Website[] }) {
  const { t } = useI18n()
  if (websites.length === 0) return <Empty text={t('No HTTP responses have been observed.')} />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>URL</TableHeader><TableHeader>{t('Status')}</TableHeader><TableHeader>{t('Title')}</TableHeader><TableHeader>{t('Technology evidence')}</TableHeader><TableHeader>{t('Last observed')}</TableHeader></TableRow></TableHead>
    <TableBody>{websites.map((website) => <TableRow key={website.id}>
      <TableCell><strong>{website.url}</strong><code>{shortId(website.content_sha256)} / {t('{count} crawled paths', { count: website.discovered_paths.length })}{website.screenshot_evidence_id ? ` / ${t('screenshot retained')}` : ''}</code></TableCell>
      <TableCell><Tag type={website.status_code < 400 ? 'green' : 'red'}>{website.status_code}</Tag></TableCell>
      <TableCell>{website.title || t('Untitled response')}</TableCell>
      <TableCell>{website.fingerprints.length > 0
        ? website.fingerprints.map((fingerprint) => <Tag key={fingerprint.id} type="blue">{fingerprint.name}{fingerprint.version ? ` ${fingerprint.version}` : ''}</Tag>)
        : website.server || t('Not identified')}</TableCell>
      <TableCell>{formatTime(website.last_seen_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function FindingTable({ findings, assets, onSelect }: { findings: Finding[]; assets: Asset[]; onSelect?: (finding: Finding) => void }) {
  const { t } = useI18n()
  const assetNames = new Map(assets.map((asset) => [asset.id, asset.value]))
  if (findings.length === 0) return <Empty text={t('No evidence-backed findings have been reported.')} />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>{t('Finding')}</TableHeader><TableHeader>{t('Asset')}</TableHeader><TableHeader>{t('Severity')}</TableHeader><TableHeader>{t('Detector')}</TableHeader><TableHeader>{t('Last observed')}</TableHeader></TableRow></TableHead>
    <TableBody>{findings.map((finding) => <TableRow key={finding.id}>
      <TableCell>{onSelect ? <Button kind="ghost" size="sm" onClick={() => onSelect(finding)}>{localizeFindingTitle(finding.title, t)}</Button> : <strong>{localizeFindingTitle(finding.title, t)}</strong>}<code>{finding.rule_id}</code></TableCell>
      <TableCell>{assetNames.get(finding.asset_id) ?? finding.asset_id}</TableCell>
      <TableCell><Tag type={finding.severity >= 4 ? 'red' : finding.severity === 3 ? 'magenta' : 'cool-gray'}>{severityName(finding.severity, t)}</Tag></TableCell>
      <TableCell>{finding.detector}</TableCell>
      <TableCell>{formatTime(finding.last_seen_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function ScheduleTable({ schedules }: { schedules: Schedule[] }) {
  const { t } = useI18n()
  if (schedules.length === 0) return <Empty text={t('No recurring discovery schedules have been recorded.')} />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>{t('Schedule')}</TableHeader><TableHeader>{t('Policy')}</TableHeader><TableHeader>{t('Cadence')}</TableHeader><TableHeader>{t('Next run')}</TableHeader><TableHeader>{t('Last task')}</TableHeader></TableRow></TableHead>
    <TableBody>{schedules.map((schedule) => <TableRow key={schedule.id}>
      <TableCell><code>{schedule.id}</code></TableCell>
      <TableCell>{t(schedule.policy_id)}</TableCell>
      <TableCell><Tag type={schedule.enabled ? 'green' : 'gray'}>{schedule.enabled ? t('Every {interval}', { interval: formatInterval(schedule.interval_seconds) }) : t('Disabled')}</Tag></TableCell>
      <TableCell>{formatTime(schedule.next_run_at)}</TableCell>
      <TableCell><code>{schedule.last_task_id ? shortId(schedule.last_task_id) : t('Not run')}</code></TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function ChangeTable({ changes, assets }: { changes: AssetChange[]; assets: Asset[] }) {
  const { t } = useI18n()
  const assetNames = new Map(assets.map((asset) => [asset.id, asset.value]))
  if (changes.length === 0) return <Empty text={t('No drift has been detected after a completed monitoring baseline.')} />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>{t('Change')}</TableHeader><TableHeader>{t('Asset')}</TableHeader><TableHeader>{t('Schedule')}</TableHeader><TableHeader>{t('Detected')}</TableHeader></TableRow></TableHead>
    <TableBody>{changes.map((change) => <TableRow key={change.id}>
      <TableCell><Tag type={change.kind === 1 ? 'green' : 'red'}>{change.kind === 1 ? t('Appeared') : t('Disappeared')}</Tag></TableCell>
      <TableCell><strong>{assetNames.get(change.asset_id) ?? t('Retained asset')}</strong><code>{change.asset_id}</code></TableCell>
      <TableCell><code>{shortId(change.schedule_id)}</code></TableCell>
      <TableCell>{formatTime(change.detected_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function ExposureChangeTable({ changes }: { changes: ExposureChange[] }) {
  const { t } = useI18n()
  if (changes.length === 0) return <Empty text={t('No service or website changes have been detected.')} />
  const labels: Record<number, string> = { 1: t('Appeared'), 2: t('Disappeared'), 3: t('Modified') }
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>{t('Resource')}</TableHeader><TableHeader>{t('Change')}</TableHeader><TableHeader>{t('Fingerprint')}</TableHeader><TableHeader>{t('Detected')}</TableHeader></TableRow></TableHead>
    <TableBody>{changes.map((change) => <TableRow key={change.id}>
      <TableCell><strong>{change.resource_kind}</strong><code>{change.resource_id}</code></TableCell>
      <TableCell><Tag type={change.kind === 2 ? 'red' : change.kind === 3 ? 'blue' : 'green'}>{labels[change.kind] ?? t('Unknown')}</Tag></TableCell>
      <TableCell><code>{shortId(change.current_fingerprint || change.previous_fingerprint)}</code></TableCell>
      <TableCell>{formatTime(change.detected_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function AuditTable({ events }: { events: AuditEvent[] }) {
  const { t } = useI18n()
  if (events.length === 0) return <Empty text={t('No Agent mutations have been audited.')} />
  return <TableContainer><Table size="lg">
    <TableHead><TableRow><TableHeader>{t('Operation')}</TableHeader><TableHeader>{t('Agent and Skill')}</TableHeader><TableHeader>{t('Resource')}</TableHeader><TableHeader>{t('Occurred')}</TableHeader></TableRow></TableHead>
    <TableBody>{events.map((event) => <TableRow key={event.id}>
      <TableCell><strong>{t(event.operation)}</strong>{event.request_id && <code>{event.request_id}</code>}</TableCell>
      <TableCell>{event.agent_id ?? t('Restricted')}{event.skill_name && <code>{event.skill_name}@{event.skill_version}</code>}</TableCell>
      <TableCell><code>{event.resource_id}</code></TableCell>
      <TableCell>{formatTime(event.occurred_at)}</TableCell>
    </TableRow>)}</TableBody>
  </Table></TableContainer>
}

function InspectorPanel({ inspector, data, observations, onClose, onEvidence }: { inspector: Inspector; data: Overview; observations: Observation[]; onClose: () => void; onEvidence: (id: string) => void }) {
  const { t } = useI18n()
  let title = t('Details')
  let content: React.ReactNode
  if (inspector.kind === 'asset') {
    const services = data.services.filter((item) => item.asset_id === inspector.asset.id)
    const serviceIds = new Set(services.map((item) => item.id))
    const findings = data.findings.filter((item) => item.asset_id === inspector.asset.id)
    title = inspector.asset.value
    content = <><DefinitionList values={{ [t('Asset ID')]: inspector.asset.id, [t('Scope ID')]: inspector.asset.scope_id, [t('Kind')]: inspector.asset.kind === 1 ? t('Domain') : t('IP address'), [t('First observed')]: formatTime(inspector.asset.first_seen_at), [t('Last observed')]: formatTime(inspector.asset.last_seen_at) }} /><h3>{t('Relationships')}</h3><p>{t('{services} services, {websites} websites, {certificates} certificates, {findings} findings', { services: services.length, websites: data.websites.filter((item) => serviceIds.has(item.service_id)).length, certificates: data.certificates.filter((item) => serviceIds.has(item.service_id)).length, findings: findings.length })}</p><FindingTable findings={findings} assets={data.assets} onSelect={(finding) => onEvidence(finding.evidence_id)} /></>
  } else if (inspector.kind === 'task') {
    title = t('Task {id}', { id: shortId(inspector.task.id) })
    content = <><DefinitionList values={{ [t('Task ID')]: inspector.task.id, [t('Policy')]: t(inspector.task.policy_id), [t('State')]: taskStateMeta(inspector.task.state, t).label, [t('Scope ID')]: inspector.task.scope_id, [t('Updated')]: formatTime(inspector.task.updated_at) }} /><h3>{t('Observation timeline')}</h3>{observations.length === 0 ? <Empty text={t('No observations are available for this task.')} /> : <ol className="timeline">{observations.map((item) => <li key={item.id}><time>{formatTime(item.observed_at)}</time><strong>{t(item.type)}</strong><code>{item.asset_id}</code><Button kind="ghost" size="sm" onClick={() => onEvidence(item.evidence_id)}>{t('Open evidence')}</Button></li>)}</ol>}</>
  } else if (inspector.kind === 'finding') {
    const finding = inspector.finding
    title = localizeFindingTitle(finding.title, t)
    content = <><DefinitionList values={{ [t('Finding ID')]: finding.id, [t('Rule')]: finding.rule_id, [t('Detector')]: finding.detector, [t('Severity')]: severityName(finding.severity, t), [t('State')]: finding.state === 1 ? t('Open') : t('Resolved'), [t('Asset ID')]: finding.asset_id, [t('Last observed')]: formatTime(finding.last_seen_at) }} /><p className="inspector-description">{localizeFindingDescription(finding.description, t)}</p><Button kind="secondary" onClick={() => onEvidence(finding.evidence_id)}>{t('Inspect supporting evidence')}</Button></>
  } else {
    title = t('Evidence {id}', { id: shortId(inspector.evidence.id) })
    content = <><DefinitionList values={{ [t('Evidence ID')]: inspector.evidence.id, [t('Media type')]: inspector.evidence.media_type, 'SHA-256': inspector.evidence.sha256, [t('Created')]: formatTime(inspector.evidence.created_at) }} /><h3>{t('Preview')}</h3><pre className="evidence-preview">{evidencePreview(inspector.evidence, t)}</pre></>
  }
  return <aside className="inspector" aria-labelledby="inspector-title" tabIndex={-1}><div className="inspector-header"><h2 id="inspector-title">{title}</h2><Button hasIconOnly kind="ghost" iconDescription={t('Close details')} renderIcon={Close} onClick={onClose} /></div><div className="inspector-body">{content}</div></aside>
}

function DefinitionList({ values }: { values: Record<string, string | number> }) {
  return <dl className="definition-list">{Object.entries(values).map(([label, value]) => <div key={label}><dt>{label}</dt><dd>{value}</dd></div>)}</dl>
}

function Loading() {
  const { t } = useI18n()
  return <div className="loading" aria-label={t('Loading read model')}><SkeletonText heading width="35%" /><SkeletonText paragraph lineCount={6} /></div>
}

function Empty({ text }: { text: string }) {
  const { t } = useI18n()
  return <div className="empty"><strong>{t('Nothing to display')}</strong><p>{text}</p></div>
}

function parseView(hash: string): View {
  const value = hash.replace('#', '') as View
  return navigation.some((item) => item.id === value) ? value : 'overview'
}

function viewLabel(view: View, t: Translate) {
  return t(({ overview: 'Overview', inventory: 'Inventory', risk: 'Findings', tasks: 'Tasks', monitoring: 'Monitoring', audit: 'Audit' } as Record<View, string>)[view])
}

function viewDescription(view: View, t: Translate) {
  const descriptions: Record<View, string> = {
    overview: 'Current exposure posture, authorization coverage, and evidence retention.',
    inventory: 'Connected assets, services, certificates, and websites from authorized scopes.',
    risk: 'Evidence-backed findings with severity and lifecycle state.',
    tasks: 'Deterministic execution history and observation timelines.',
    monitoring: 'Recurring schedules, exposure drift, and notification delivery.',
    audit: 'Agent and Skill provenance for every control-plane mutation.',
  }
  return t(descriptions[view])
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

function localizeFindingTitle(title: string, t: Translate) {
  const cve = title.match(/^(CVE-\d{4}-\d+) applies to observed product identity$/)
  return cve ? t('{cve} applies to observed product identity', { cve: cve[1] }) : t(title)
}

function localizeFindingDescription(description: string, t: Translate) {
  const matched = description.match(/^Matched at (.+)\.$/)
  if (matched) return t('Matched at {target}.', { target: matched[1] })

  const collision = description.match(/^Sending the authorized Host (.+) to (.+) returned a materially different successful response than the address baseline\.$/)
  if (collision) return t('Sending the authorized Host {host} to {address} returned a materially different successful response than the address baseline.', { host: collision[1], address: collision[2] })

  const certificate = description.match(/^The retained DER certificate for service (.+) expires at Unix timestamp (.+)\.$/)
  if (certificate) return t('The retained DER certificate for service {service} expires at Unix timestamp {timestamp}.', { service: certificate[1], timestamp: certificate[2] })

  const fixedProbe = description.match(/^The fixed HTTP probe at (.+) returned content matching (.+)\.$/)
  if (fixedProbe) return t('The fixed HTTP probe at {url} returned content matching {rule}.', { url: fixedProbe[1], rule: fixedProbe[2] })

  const directory = description.match(/^The HTTP response at (.+) exposes a directory index backed by the retained response body\.$/)
  if (directory) return t('The HTTP response at {url} exposes a directory index backed by the retained response body.', { url: directory[1] })

  return t(description)
}

function evidencePreview(evidence: Evidence, t: Translate) {
  if (!evidence.media_type.startsWith('text/') && !evidence.media_type.includes('json')) return t('Binary evidence preview is unavailable. Verify it using the retained digest.')
  try {
    const bytes = Uint8Array.from(atob(evidence.content_base64), (value) => value.charCodeAt(0))
    const value = new TextDecoder().decode(bytes.slice(0, 32768))
    return bytes.length > 32768 ? `${value}\n${t('[Preview truncated at 32 KiB]')}` : value
  } catch {
    return t('Evidence content could not be decoded.')
  }
}

function taskStateMeta(value: number, t: Translate): { label: string; type: 'gray' | 'blue' | 'green' | 'red' } {
  return ({
    1: { label: t('Queued'), type: 'gray' },
    2: { label: t('Running'), type: 'blue' },
    3: { label: t('Completed'), type: 'green' },
    4: { label: t('Failed'), type: 'red' },
    5: { label: t('Canceled'), type: 'gray' },
  } as Record<number, { label: string; type: 'gray' | 'blue' | 'green' | 'red' }>)[value] ?? { label: t('Unknown'), type: 'gray' }
}

function severityName(value: number, t: Translate) {
  return t(({ 1: 'Info', 2: 'Low', 3: 'Medium', 4: 'High', 5: 'Critical' } as Record<number, string>)[value] ?? 'Unknown')
}

function shortId(value: string) { return value.length > 22 ? `${value.slice(0, 18)}...` : value }
function formatNumber(value: number) { return new Intl.NumberFormat(document.documentElement.lang === 'zh-CN' ? 'zh-CN' : 'en-US').format(value) }
function formatInterval(seconds: number) {
  if (seconds % 86400 === 0) return `${seconds / 86400}d`
  if (seconds % 3600 === 0) return `${seconds / 3600}h`
  if (seconds % 60 === 0) return `${seconds / 60}m`
  return `${seconds}s`
}
function formatTime(stamp: Stamp) {
  if (!stamp) return document.documentElement.lang === 'zh-CN' ? '未知' : 'Unknown'
  return new Intl.DateTimeFormat(document.documentElement.lang === 'zh-CN' ? 'zh-CN' : 'en-GB', { dateStyle: 'medium', timeStyle: 'medium' }).format(new Date(stamp.seconds * 1000))
}

const en: Record<string, string> = {
  'policy_passive_dns': 'Passive DNS discovery', 'policy_passive_inventory': 'Passive inventory',
  'policy_service_baseline': 'Service baseline scan', 'policy_vulnerability_baseline': 'Vulnerability baseline scan',
  'policy_public_code_intelligence': 'Public code intelligence', 'policy_cve_intelligence': 'CVE intelligence',
  'policy_registration_intelligence': 'Registration intelligence', 'policy_assessment_standard': 'Standard assessment',
  'policy_assessment_thorough': 'Thorough assessment',
  'scope.create': 'Create scope', 'scan.start': 'Start scan', 'assessment.start': 'Start assessment',
  'schedule.create': 'Create schedule', 'task.cancel': 'Cancel task', 'finding.report': 'Report finding',
  'assessment.coverage': 'Assessment coverage', 'ct.discovered_domain': 'CT discovered domain', 'ct.error': 'CT error',
  'dns.addresses': 'DNS addresses', 'dns.discovered_ip': 'DNS discovered IP', 'dns.error': 'DNS error', 'dns.wildcard': 'DNS wildcard',
  'github.code.coverage': 'GitHub code coverage', 'github.code.error': 'GitHub code error', 'github.code.reference': 'GitHub code reference',
  'http.crawl': 'HTTP crawl', 'http.crawl_error': 'HTTP crawl error', 'http.error': 'HTTP error',
  'http.exposure_probe': 'HTTP exposure probe', 'http.host_collision_check': 'HTTP Host collision check',
  'http.host_collision_error': 'HTTP Host collision error', 'http.response': 'HTTP response',
  'http.screenshot': 'HTTP screenshot', 'http.screenshot_error': 'HTTP screenshot error',
  'nuclei.coverage': 'Nuclei coverage', 'nuclei.error': 'Nuclei error', 'nuclei.result': 'Nuclei result',
  'nvd.cve.result': 'NVD CVE result', 'nvd.cve.no_targets': 'NVD CVE has no targets',
  'registration.error': 'Registration intelligence error', 'tcp.no_open_ports': 'No open TCP ports',
  'tcp.open': 'Open TCP port', 'tls.certificate': 'TLS certificate', 'tls.error': 'TLS error',
}

const zh: Record<string, string> = {
  'CyberEdge Observer': 'CyberEdge 观察台',
  'Overview': '概览', 'Inventory': '资产清单', 'Findings': '风险发现', 'Tasks': '任务', 'Monitoring': '监控', 'Audit': '审计',
  'Queued': '排队中', 'Running': '运行中', 'Completed': '已完成', 'Failed': '失败', 'Canceled': '已取消', 'Unknown': '未知',
  'Info': '信息', 'Low': '低危', 'Medium': '中危', 'High': '高危', 'Critical': '严重',
  'Skip to main content': '跳转到主要内容', 'CyberEdge read-only observer': 'CyberEdge 只读观察台', 'READ ONLY': '只读',
  'Switch language': '切换语言', 'Observation navigation': '观察台导航', 'Mobile observation navigation': '移动端观察台导航',
  'External attack surface': '外部攻击面', 'Refresh': '刷新', 'Save view': '保存视图', 'Export JSON': '导出 JSON',
  'Search all projections': '搜索全部数据', 'Search assets, findings, tasks, certificates, audit': '搜索资产、风险、任务、证书和审计记录',
  'Severity': '严重程度', 'State': '状态', 'All': '全部', 'Saved view': '已保存视图', 'Choose': '选择',
  'Name this local view': '为此本地视图命名', 'Read model unavailable': '只读数据不可用',
  'Read model returned {status}': '只读数据返回状态码 {status}', 'Task observations returned {status}': '任务观察记录返回状态码 {status}',
  'Evidence access returned {status}': 'Evidence 访问返回状态码 {status}',
  'Authorized scopes': '授权范围', 'Observed assets': '已观测资产', 'Open findings': '未关闭风险', 'Evidence objects': 'Evidence 数量',
  'Authorization boundary': '授权边界', 'Scopes': '范围', 'No authorized scopes have been recorded.': '尚未记录授权范围。',
  'Chain of custody': '证据链', 'Evidence retention': 'Evidence 留存',
  'Content-addressed JSON and binary evidence objects are linked from observations. Mutations are unavailable on this surface.': '内容寻址的 JSON 与二进制 Evidence 均关联到观察记录。此界面不提供任何修改操作。',
  'Highest priority findings': '最高优先级风险', 'View all': '查看全部', 'Assets': '资产', 'Services': '服务', 'Certificates': '证书', 'Websites': '网站',
  'Evidence-backed findings': '有 Evidence 支撑的风险', 'Task history': '任务历史', 'Monitoring schedules': '监控计划',
  'Asset changes': '资产变化', 'Exposure changes': '暴露面变化', 'Invocation audit': '调用审计',
  '{count} visible': '显示 {count} 项', '{count} recorded': '已记录 {count} 项',
  'Notifications: {delivered} delivered, {pending} pending, {dead} dead-lettered': '通知：已送达 {delivered}，等待中 {pending}，死信 {dead}',
  'Asset': '资产', 'Kind': '类型', 'Scope': '范围', 'Last observed': '最后观测时间', 'Domain': '域名', 'IP address': 'IP 地址', 'Other': '其他',
  'No assets match the current read-only view.': '当前只读视图中没有匹配的资产。',
  'Task': '任务', 'Policy': '策略', 'Updated': '更新时间', 'No discovery tasks have been recorded.': '尚未记录发现任务。',
  'Endpoint': '端点', 'Service hint': '服务提示', 'Retained asset': '留存资产',
  'No open baseline TCP services have been observed.': '尚未观测到开放的基线 TCP 服务。',
  'Subject': '主体', 'Issuer': '签发者', 'Expires': '到期时间', 'Retained service': '留存服务', 'No DNS SAN': '无 DNS SAN',
  'No TLS certificates have been observed.': '尚未观测到 TLS 证书。',
  'Status': '状态', 'Title': '标题', 'Technology evidence': '技术 Evidence', 'Untitled response': '无标题响应', 'Not identified': '未识别',
  'No HTTP responses have been observed.': '尚未观测到 HTTP 响应。', '{count} crawled paths': '爬取 {count} 个路径', 'screenshot retained': '截图已留存',
  'Finding': '风险', 'Detector': '检测器', 'No evidence-backed findings have been reported.': '尚未报告有 Evidence 支撑的风险。',
  'Schedule': '计划', 'Cadence': '周期', 'Next run': '下次运行', 'Last task': '上次任务', 'Disabled': '已禁用', 'Not run': '尚未运行',
  'Every {interval}': '每 {interval}', 'No recurring discovery schedules have been recorded.': '尚未记录周期性发现计划。',
  'Change': '变化', 'Detected': '检测时间', 'Appeared': '新增', 'Disappeared': '消失', 'Modified': '已修改',
  'No drift has been detected after a completed monitoring baseline.': '完成监控基线后尚未检测到漂移。',
  'Resource': '资源', 'Fingerprint': '指纹', 'No service or website changes have been detected.': '尚未检测到服务或网站变化。',
  'Operation': '操作', 'Agent and Skill': 'Agent 与 Skill', 'Occurred': '发生时间', 'Restricted': '受限',
  'No Agent mutations have been audited.': '尚未审计到 Agent 修改操作。',
  'Details': '详情', 'Asset ID': '资产 ID', 'Scope ID': '范围 ID', 'First observed': '首次观测时间', 'Relationships': '关联关系',
  '{services} services, {websites} websites, {certificates} certificates, {findings} findings': '{services} 个服务，{websites} 个网站，{certificates} 张证书，{findings} 条风险',
  'Task {id}': '任务 {id}', 'Task ID': '任务 ID', 'Observation timeline': '观察时间线', 'No observations are available for this task.': '此任务没有可用的观察记录。',
  'Open evidence': '打开 Evidence', 'Finding ID': '风险 ID', 'Rule': '规则', 'Open': '未关闭', 'Resolved': '已解决',
  'Inspect supporting evidence': '查看支撑 Evidence', 'Evidence {id}': 'Evidence {id}', 'Evidence ID': 'Evidence ID', 'Media type': '媒体类型',
  'Created': '创建时间', 'Preview': '预览', 'Close details': '关闭详情', 'Loading read model': '正在加载只读数据',
  'Nothing to display': '暂无数据', 'Binary evidence preview is unavailable. Verify it using the retained digest.': '二进制 Evidence 无法预览，请使用留存摘要进行验证。',
  '[Preview truncated at 32 KiB]': '[预览已在 32 KiB 处截断]', 'Evidence content could not be decoded.': '无法解码 Evidence 内容。',
  'Current exposure posture, authorization coverage, and evidence retention.': '当前暴露态势、授权覆盖与 Evidence 留存情况。',
  'Connected assets, services, certificates, and websites from authorized scopes.': '授权范围内相互关联的资产、服务、证书与网站。',
  'Evidence-backed findings with severity and lifecycle state.': '包含严重程度与生命周期状态、由 Evidence 支撑的风险。',
  'Deterministic execution history and observation timelines.': '确定性的执行历史与观察时间线。',
  'Recurring schedules, exposure drift, and notification delivery.': '周期计划、暴露面漂移与通知投递状态。',
  'Agent and Skill provenance for every control-plane mutation.': '每次控制面修改对应的 Agent 与 Skill 来源。',
  'Persistent scope': '持久化范围', 'Yuzhian Company Domain': '宇展公司域名',
  'policy_passive_dns': '被动 DNS 发现', 'policy_passive_inventory': '被动资产清单',
  'policy_service_baseline': '服务基线扫描', 'policy_vulnerability_baseline': '漏洞基线扫描',
  'policy_public_code_intelligence': '公开代码情报', 'policy_cve_intelligence': 'CVE 情报',
  'policy_registration_intelligence': '备案情报', 'policy_assessment_standard': '标准评估', 'policy_assessment_thorough': '深度评估',
  'scope.create': '创建范围', 'scan.start': '启动扫描', 'assessment.start': '启动评估',
  'schedule.create': '创建计划', 'task.cancel': '取消任务', 'finding.report': '报告风险',
  'Host-routed website exposed on alternate scoped IP': '在其他范围内 IP 上暴露了 Host 路由网站',
  'Public code references an authorized domain': '公开代码引用了授权域名',
  'Detect SSL Certificate Issuer': '检测 SSL 证书签发者', 'SSL DNS Names': 'SSL DNS 名称',
  'TLS Version - Detect': 'TLS 版本检测', 'Wildcard TLS Certificate': '通配符 TLS 证书',
  'CORS Misconfiguration': 'CORS 配置错误', 'TLS certificate expired': 'TLS 证书已过期',
  'TLS certificate expires within 30 days': 'TLS 证书将在 30 天内到期',
  'Git repository metadata exposed': 'Git 仓库元数据暴露', 'DS_Store metadata exposed': 'DS_Store 元数据暴露',
  'HTTP directory listing exposed': 'HTTP 目录列表暴露',
  '{cve} applies to observed product identity': '{cve} 与观测到的产品身份相关',
  'GitHub code search returned a public repository path referencing the authorized domain. This is a review candidate, not proof of a credential leak.': 'GitHub 代码搜索发现公开仓库路径引用了授权域名。此结果仅供复核，不代表已经发生凭据泄漏。',
  'Matched at {target}.': '命中位置：{target}。',
  'Sending the authorized Host {host} to {address} returned a materially different successful response than the address baseline.': '将授权 Host {host} 发送到 {address} 后，返回了与地址基线明显不同的成功响应。',
  'The retained DER certificate for service {service} expires at Unix timestamp {timestamp}.': '服务 {service} 留存的 DER 证书将在 Unix 时间戳 {timestamp} 到期。',
  'The fixed HTTP probe at {url} returned content matching {rule}.': '固定 HTTP 探针在 {url} 返回了匹配规则 {rule} 的内容。',
  'The HTTP response at {url} exposes a directory index backed by the retained response body.': '{url} 的 HTTP 响应暴露了目录索引，留存的响应正文可作为 Evidence。',
  'assessment.coverage': '评估覆盖', 'ct.discovered_domain': 'CT 发现域名', 'ct.error': 'CT 错误',
  'dns.addresses': 'DNS 地址', 'dns.discovered_ip': 'DNS 发现 IP', 'dns.error': 'DNS 错误', 'dns.wildcard': 'DNS 通配符',
  'github.code.coverage': 'GitHub 代码覆盖', 'github.code.error': 'GitHub 代码错误', 'github.code.reference': 'GitHub 代码引用',
  'http.crawl': 'HTTP 爬取', 'http.crawl_error': 'HTTP 爬取错误', 'http.error': 'HTTP 错误',
  'http.exposure_probe': 'HTTP 暴露探测', 'http.host_collision_check': 'HTTP Host 冲突检查',
  'http.host_collision_error': 'HTTP Host 冲突检查错误', 'http.response': 'HTTP 响应',
  'http.screenshot': 'HTTP 截图', 'http.screenshot_error': 'HTTP 截图错误',
  'nuclei.coverage': 'Nuclei 覆盖', 'nuclei.error': 'Nuclei 错误', 'nuclei.result': 'Nuclei 结果',
  'nvd.cve.result': 'NVD CVE 结果', 'nvd.cve.no_targets': 'NVD CVE 无目标',
  'registration.error': '备案情报错误', 'tcp.no_open_ports': 'TCP 无开放端口', 'tcp.open': 'TCP 端口开放',
  'tls.certificate': 'TLS 证书', 'tls.error': 'TLS 错误',
}

const LocaleContext = createContext<{ locale: Locale; setLocale: (locale: Locale) => void; t: Translate } | null>(null)

function I18nRoot() {
  const [locale, setLocaleState] = useState<Locale>(() => {
    const saved = localStorage.getItem('cyberedge.locale')
    return saved === 'en' || saved === 'zh-CN' ? saved : navigator.language.toLowerCase().startsWith('zh') ? 'zh-CN' : 'en'
  })
  const setLocale = (next: Locale) => {
    localStorage.setItem('cyberedge.locale', next)
    setLocaleState(next)
  }
  const t = useMemo<Translate>(() => (key, values = {}) => {
    const template = locale === 'zh-CN' ? zh[key] ?? key : en[key] ?? key
    return Object.entries(values).reduce((value, [name, replacement]) => value.replaceAll(`{${name}}`, String(replacement)), template)
  }, [locale])
  useEffect(() => {
    document.documentElement.lang = locale
    document.title = locale === 'zh-CN' ? 'CyberEdge 观察台' : 'CyberEdge Observer'
  }, [locale])
  return <LocaleContext.Provider value={{ locale, setLocale, t }}><App /></LocaleContext.Provider>
}

function useI18n() {
  const context = useContext(LocaleContext)
  if (!context) throw new Error('I18n context is unavailable')
  return context
}

createRoot(document.getElementById('root')!).render(<StrictMode><I18nRoot /></StrictMode>)
