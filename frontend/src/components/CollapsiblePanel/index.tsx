import React from 'react';
import { ChevronDown, ChevronUp } from 'lucide-react';
import styles from './index.module.css';

interface CollapsiblePanelProps {
  title: string;
  children: React.ReactNode;
  collapsed: boolean;
  onToggle: () => void;
}

const CollapsiblePanel: React.FC<CollapsiblePanelProps> = ({
  title,
  children,
  collapsed,
  onToggle,
}) => {
  return (
    <div className={`${styles.panel} ${collapsed ? styles.collapsed : styles.expanded}`}>
      <div className={styles.header} onClick={onToggle}>
        <span className={styles.title}>{title}</span>
        <button className={styles.toggleButton}>
          {collapsed ? <ChevronDown size={16} /> : <ChevronUp size={16} />}
        </button>
      </div>
      <div className={styles.content} style={{ display: collapsed ? 'none' : 'block' }}>
        {children}
      </div>
    </div>
  );
};

export default CollapsiblePanel;
