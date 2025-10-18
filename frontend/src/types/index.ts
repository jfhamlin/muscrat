export interface Buffer {
  fileName: string;
  content: string;
  dirty: boolean;
  isTemp?: boolean;
}

// Wails binding types
export interface OpenFileDialogResponse {
  FileName: string;
  Content: string;
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
  color?: string;
  onChange?: (value: number) => void;
}

export interface KnobData {
  id: number;
  name: string;
  min: number;
  max: number;
  def: number;
  step: number;
  group: string;
}

export interface EditorProps {
  [key: string]: any;
}

export interface OscilloscopeProps {
  analyser: AnalyserNode | null;
}

export interface SpectrogramProps {
  analyser: AnalyserNode | null;
  sampleRate: number;
}

export interface TabBarProps {
  options: string[];
  selected: string;
  onSelect: (option: string) => void;
}

export interface TabButtonProps {
  children: React.ReactNode;
  selected: boolean;
  onClick: () => void;
}

export interface ScopeDisplayProps {
  id: string;
  samples: Float32Array;
  sampleRate: number;
  name: string;
  width?: number;
  height?: number;
}

export interface ScopeData {
  id: string;
  samples: Float32Array;
  sampleRate: number;
  name: string;
  lastUpdate: number;
}

export interface ScopeInfo {
  id: string;
  name?: string;
}

export interface ScopeEventData {
  id: string;
  samples: Float32Array;
  sampleRate: number;
  name: string;
}

export interface ScopesChangedData {
  scopes: ScopeInfo[];
}

export interface HeadingProps {
  children: React.ReactNode;
  level?: 1 | 2 | 3 | 4 | 5 | 6;
  className?: string;
}

export interface TimerProps {
  startTime: Date | undefined;
  setStartTime: (date?: Date) => void;
}

export interface ConsoleEvent {
  level: 'debug' | 'info' | 'warn' | 'error';
  message: string;
  data?: any;
}

export interface EventWithCount {
  event: ConsoleEvent;
  count: number;
}

export interface EventProps {
  event: ConsoleEvent;
  count: number;
}

export interface ClearButtonProps {
  onClick: () => void;
  size?: number;
}

export interface VolumeData {
  rms: [number, number];
  peak: [number, number];
  rmsDB: [number, number];
  peakDB: [number, number];
}

export interface VolumeEvent {
  data: [VolumeData];
}

export interface VolumeMeterProps {
  // Currently no props, but interface ready for future props
}

export interface ButtonProps {
  children?: React.ReactNode;
  onClick?: () => void;
  disabled?: boolean;
  title?: string;
  className?: string;
}

export interface SvgProps {
  children: React.ReactNode;
}