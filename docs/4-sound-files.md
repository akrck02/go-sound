# 🎼 Sound files

| Format | Package | Read |
|---|---|---|
| WAV | `wav` | Integer samples of 8, 16, 24 and 32 bits, floating point ones of 32 and 64, any channels, any rate. Written here: a header and samples |
| Ogg Vorbis | `vorbis` | Any Vorbis file, decoded in Go by [jfreymuth/oggvorbis](https://github.com/jfreymuth/oggvorbis) |

**Which to choose.** Ogg Vorbis for music and anything longer than a second: about a tenth of the
size of a WAV, free of patents, decoded in Go everywhere. WAV for very short effects played often,
where decoding costs more than the bytes.

Both rewind when their source can seek, so a player can loop them:

```go
//go:embed waves.ogg
var waves []byte

stream, err := vorbis.Decode(bytes.NewReader(waves)) // a bytes.Reader seeks
```

**Not read yet.** MP3 (its Go decoder, go-mp3, is archived) and Opus (Go decoders exist, but are
young). A film's sound (AAC or Opus inside an MP4) will come from the system's own decoders, as its
pictures do: see [the design](5-the-design.md#films).
