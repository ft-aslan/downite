import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './page.tsx'
import RootLayout from './layout.tsx'
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import './globals.css';


ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<RootLayout/>} >
            <Route index element={<App />} />
        </Route>
      </Routes>
    </BrowserRouter>
  </React.StrictMode>,
)


