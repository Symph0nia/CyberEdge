import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './styles.css'

function App() {
  return (
    <main>
      <p className="eyebrow">CYBEREDGE</p>
      <h1>从边界看清风险。</h1>
      <p className="summary">
        新一代外部攻击面管理平台正在重建。资产发现、验证、证据与处置将围绕同一条可审计的数据链路展开。
      </p>
      <div className="status"><span /> Architecture reset complete</div>
    </main>
  )
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
