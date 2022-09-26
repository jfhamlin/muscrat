
Let's build a bespoke software synthesizer.

Step 1: Play a sine wave
- Pipe a signal to speakers. SDL library hooks exposed by the bleep library.
- Sample a sine wave

Step 2: Compose multiple sine waves.
- watch out for clipping

Step 3: Build an audio processing graph for easier composition from a GUI.
- inspired by Max/MSP
- consider requirements:
  1. signals streams should be processed in lock step
  2. scheduling should be deterministic if possible
  3. consider how configuration of signal generators can be updated
     dynamically. Max supports two ways of doing this:
     a. nodes sending messages to each other
     b. routing signals between nodes
