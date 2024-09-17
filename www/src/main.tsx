import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { LogViewerComponent } from './components/log-viewer'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <LogViewerComponent />
  </StrictMode>,
)
