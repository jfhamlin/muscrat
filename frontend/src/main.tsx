import React from 'react'
import ReactDOM from 'react-dom/client'
import './index.css'
import App from './App'

// @ts-ignore - Components not yet converted to TypeScript
import HydraView from "./components/HydraView";
// @ts-ignore
import Knobs from "./components/Knobs";

const Root: React.FC = () => {
  const pathname = window.location.pathname;
  let component: React.ReactElement;

  switch (pathname) {
    case '/hydra':
      component = <HydraView />;
      break;
    case '/knobs':
      component = <Knobs />;
      break;
    default:
      component = <App />;
      break;
  }
  return (
    <React.StrictMode>
      {component}
    </React.StrictMode>
  )
}

const rootElement = document.getElementById('root');
if (rootElement) {
  ReactDOM.createRoot(rootElement).render(<Root />);
}
