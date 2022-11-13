import Box from '@mui/material/Box';
import Slider from '@mui/material/Slider';

import LabeledSlider from '../LabeledSlider';

interface VolumeSliderProps {
  volume: number;
  onChange: (volume: number) => void;
}

function toDecibels(volume: number) {
  return Math.round(Math.log10(volume) * 20);
}

function fromDecibels(decibels: number) {
  return Math.pow(10, decibels / 20);
}

export default function VolumeSlider(props: VolumeSliderProps) {
  const handleChange = (event: React.ChangeEvent<{}>, value: number | number[]) => {
    // convert decibels to linear scale
    const volume = Math.pow(10, value / 20);
    props.onChange(volume);
  };

  return <LabeledSlider
    label="Volume"
    value={toDecibels(props.volume)}
    onChange={handleChange}
    min={-30}
    max={0}
    step={0.5}
  />;
}

