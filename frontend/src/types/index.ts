export interface Buffer {
  fileName: string;
  content: string;
  dirty: boolean;
  isTemp?: boolean;
}

export interface BuffersState {
  selectedBufferName: string | null;
  buffers: Record<string, Buffer> & { null?: Buffer };
  selectBuffer: (fileName: string | null) => void;
  addBuffer: (props: { fileName: string; content: string }) => void;
  updateBuffer: (fileName: string | null, content: string, dirty?: boolean) => void;
  cleanBuffer: (fileName: string | null, content: string) => void;
  playBuffer: (bufferKey: string | null) => Promise<void>;
}

export interface KnobProps {
  label?: string;
  value: number;
  min: number;
  max: number;
  step?: number;
  size?: number;
  onChange?: (value: number) => void;
}

export interface EditorProps {
  [key: string]: any;
}

export interface OscilloscopeProps {
  [key: string]: any;
}

export interface SpectrogramProps {
  [key: string]: any;
}

export interface TabBarProps {
  tabs: string[];
  activeTab: string;
  onTabChange: (tab: string) => void;
}

export interface ScopeDisplayProps {
  title: string;
  samples: Float32Array;
  width?: number;
  height?: number;
}

export interface HeadingProps {
  children: React.ReactNode;
  level?: 1 | 2 | 3 | 4 | 5 | 6;
  className?: string;
}