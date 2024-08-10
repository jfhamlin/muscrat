import React from 'react'
import ReactDOM from 'react-dom/client'
import './index.css'
import App from './App'

import HydraView from "./components/HydraView";
import Knobs from "./components/Knobs";

const Root = () => {
  const pathname = window.location.pathname;
  let component;

  switch (pathname) {
    case '/hydra':
      component = <HydraView />;
      break;
    case '/knobs':
      component = <Knobs />;
      break;
    default:
      component = <App />;
  }
  return (
    <React.StrictMode>
      {component}
    </React.StrictMode>
  )
}

ReactDOM.createRoot(document.getElementById('root')).render(<Root />)
