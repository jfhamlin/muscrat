import Box from '@mui/material/Box';
import Slider from '@mui/material/Slider';
import Typography from '@mui/material/Typography';

export default function LabeledSlider(props: {
  label: string;
  min: number;
  max: number;
  step: number;
  value: number;
  onChange: (event: React.ChangeEvent<{}>, newValue: number | number[]) => void;
}) {
  return (
    <Box>
      <Typography gutterBottom>{props.label}</Typography>
      <Slider
        aria-label={props.label}
        size="small"
        min={props.min}
        max={props.max}
        step={props.step}
        value={props.value}
        onChange={props.onChange}
        valueLabelDisplay="auto" />
    </Box>
  );
}
