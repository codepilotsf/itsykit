# Video test fixtures

These tiny synthetic fixtures were generated offline with ffmpeg. Flea does not
require or invoke ffmpeg at runtime.

- `valid.mp4`: 32×24 blue video, 0.3 seconds, H.264/yuv420p and AAC, faststart.
- `silent.mp4`: the same H.264 video without audio, faststart.
- `no-faststart.mp4`: the valid streams remuxed with metadata after media data.
- `wrong-codec.mp4`: MPEG-4 Part 2 video, faststart, no audio.
- `wrong-audio.mp4`: H.264 video with MP3 rather than AAC audio, faststart.

Generation used a synthetic color source and sine wave, with no third-party media.
