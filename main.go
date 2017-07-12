package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
)

const (
	SampleRate   = 44100
	NoteDuration = 200 * time.Millisecond
	BPM          = 120 // BPM
)

var songs = map[string][]Note{
	"bumblebees": []Note{
		{E6, d16}, {Ef6, d16}, {D6, d16}, {Df6, d16}, {D6, d16}, {Df6, d16}, {C6, d16}, {B5, d16},
		{C6, d16}, {B5, d16}, {Bf5, d16}, {A5, d16}, {Af5, d16}, {G5, d16}, {Gf5, d16}, {F5, d16},
		{E5, d16}, {Ef5, d16}, {D5, d16}, {Df5, d16}, {D5, d16}, {Df5, d16}, {C5, d16}, {B4, d16},
		{C5, d16}, {B4, d16}, {Bf4, d16}, {A4, d16}, {Af4, d16}, {G4, d16}, {Gf4, d16}, {F4, d16},
		{E4, d16}, {Ef4, d16}, {D4, d16}, {Df4, d16}, {D4, d16}, {Df4, d16}, {C4, d16}, {B3, d16},
		{E4, d16}, {Ef4, d16}, {D4, d16}, {Df4, d16}, {D4, d16}, {Df4, d16}, {C4, d16}, {B3, d16},
		{E4, d16}, {Ef4, d16}, {D4, d16}, {Df4, d16}, {C4, d16}, {F4, d16}, {E4, d16}, {Ef4, d16},
		{E4, d16}, {Ef4, d16}, {D4, d16}, {Df4, d16}, {C4, d16}, {Df4, d16}, {D4, d16}, {Ds4, d16},
	},
}

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()

	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("hackerbeeper", flag.ContinueOnError)
	song := flag.String("song", "bumblebees", "song name")
	if err := fs.Parse(args); err != nil {
		return err
	}

	// Find the song.
	notes := songs[*song]
	if notes == nil {
		return errors.New("song not found")
	}

	// Execute command.
	switch fs.Arg(0) {
	case "":
		return playWithKeyboard(notes)
	case "autoplay":
		return autoplay(notes)
	default:
		return errors.New("unknown command")
	}
}

func autoplay(notes []Note) error {
	stream, err := StartDefaultStream()
	if err != nil {
		return err
	}
	defer stream.Close()

	for _, note := range notes {
		stream.SetFrequency(note.Frequency)
		time.Sleep(note.Duration)
	}
	stream.SetFrequency(0)
	return nil
}

func playWithKeyboard(notes []Note) error {
	if err := exec.Command("stty", "-f", "/dev/tty", "cbreak", "min", "1").Run(); err != nil {
		return err
	}

	stream, err := StartDefaultStream()
	if err != nil {
		return err
	}
	defer stream.Close()

	for i := 0; ; i = (i + 1) % len(notes) {
		os.Stdin.Read(make([]byte, 1))
		stream.PlayNote(notes[i])
	}

	return nil
}

type Stream struct {
	*portaudio.Stream

	mu    sync.Mutex
	freq  float64
	timer *time.Timer

	n int64
}

func StartDefaultStream() (*Stream, error) {
	stream := &Stream{}
	s, err := portaudio.OpenDefaultStream(0, 1, SampleRate, 0, stream.callback)
	if err != nil {
		return nil, err
	}
	stream.Stream = s

	if err := stream.Start(); err != nil {
		return nil, err
	}
	return stream, nil
}

func (s *Stream) Frequency() float64 {
	s.mu.Lock()
	v := s.freq
	s.mu.Unlock()
	return v
}

func (s *Stream) SetFrequency(v float64) {
	s.mu.Lock()
	s.freq = v
	s.mu.Unlock()
}

func (s *Stream) PlayNote(note Note) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.freq = note.Frequency

	if s.timer != nil {
		s.timer.Stop()
	}
	s.timer = time.AfterFunc(note.Duration, s.Silence)
}

func (s *Stream) Silence() {
	s.SetFrequency(0)
}

func (s *Stream) callback(out [][]float32) {
	freq := s.Frequency()
	for i := range out[0] {
		out[0][i] = float32(math.Sin(freq * 2 * math.Pi * (float64(s.n) / float64(SampleRate))))
		s.n++
	}
}

type Note struct {
	Frequency float64
	Duration  time.Duration
}

// Note durations.
const (
	d1  = 4 * (time.Minute / BPM)
	d2  = 2 * (time.Minute / BPM)
	d4  = 1 * (time.Minute / BPM)
	d8  = (time.Minute / BPM) / 2
	d16 = (time.Minute / BPM) / 4
)

// Note frequencies.
const (
	C0  = 32.7032
	Cs0 = 34.6478
	Df0 = 34.6478
	D0  = 36.7081
	Ds0 = 38.8909
	Ef0 = 38.8909
	E0  = 41.2034
	F0  = 43.6535
	Fs0 = 46.2493
	Gf0 = 46.2493
	G0  = 48.9994
	Gs0 = 51.9131
	Af0 = 51.9131
	A0  = 55.0000
	As0 = 58.2705
	Bf0 = 58.2705
	B0  = 61.7354
	C1  = 65.4064
	Cs1 = 69.2957
	Df1 = 69.2957
	D1  = 73.4162
	Ds1 = 77.7817
	Ef1 = 77.7817
	E1  = 82.4069
	F1  = 87.3071
	Fs1 = 92.4986
	Gf1 = 92.4986
	G1  = 97.9989
	Gs1 = 103.8262
	Af1 = 103.8262
	A1  = 110.0000
	As1 = 116.5409
	Bf1 = 116.5409
	B1  = 123.4708
	C2  = 130.8128
	Cs2 = 138.5913
	Df2 = 138.5913
	D2  = 146.8324
	Ds2 = 155.5635
	Ef2 = 155.5635
	E2  = 164.8138
	F2  = 174.6141
	Fs2 = 184.9972
	Gf2 = 184.9972
	G2  = 195.9977
	Gs2 = 207.6523
	Af2 = 207.6523
	A2  = 220.0000
	As2 = 233.0819
	Bf2 = 233.0819
	B2  = 246.9417
	C3  = 261.6256
	Cs3 = 277.1826
	Df3 = 277.1826
	D3  = 293.6648
	Ds3 = 311.1270
	Ef3 = 311.1270
	E3  = 329.6276
	F3  = 349.2282
	Fs3 = 369.9944
	Gf3 = 369.9944
	G3  = 391.9954
	Gs3 = 415.3047
	Af3 = 415.3047
	A3  = 440.0000
	As3 = 466.1638
	Bf3 = 466.1638
	B3  = 493.8833
	C4  = 523.2511
	Cs4 = 554.3653
	Df4 = 554.3653
	D4  = 587.3295
	Ds4 = 622.2540
	Ef4 = 622.2540
	E4  = 659.2551
	F4  = 698.4565
	Fs4 = 739.9888
	Gf4 = 739.9888
	G4  = 783.9909
	Gs4 = 830.6094
	Af4 = 830.6094
	A4  = 880.0000
	As4 = 932.3275
	Bf4 = 932.3275
	B4  = 987.7666
	C5  = 1046.5023
	Cs5 = 1108.7305
	Df5 = 1108.7305
	D5  = 1174.6591
	Ds5 = 1244.5079
	Ef5 = 1244.5079
	E5  = 1318.5102
	F5  = 1396.9129
	Fs5 = 1479.9777
	Gf5 = 1479.9777
	G5  = 1567.9817
	Gs5 = 1661.2188
	Af5 = 1661.2188
	A5  = 1760.0000
	As5 = 1864.6550
	Bf5 = 1864.6550
	B5  = 1975.5332
	C6  = 2093.0045
	Cs6 = 2217.4610
	Df6 = 2217.4610
	D6  = 2349.3181
	Ds6 = 2489.0159
	Ef6 = 2489.0159
	E6  = 2637.0205
	F6  = 2793.8259
	Fs6 = 2959.9554
	Gf6 = 2959.9554
	G6  = 3135.9635
	Gs6 = 3322.4376
	Af6 = 3322.4376
	A6  = 3520.0000
	As6 = 3729.3101
	Bf6 = 3729.3101
	B6  = 3951.0664
	C7  = 4186.0090
	Cs7 = 4434.9221
	Df7 = 4434.9221
	D7  = 4698.6363
	Ds7 = 4978.0317
	Ef7 = 4978.0317
	E7  = 5274.0409
	F7  = 5587.6517
	Fs7 = 5919.9108
	Gf7 = 5919.9108
	G7  = 6271.9270
	Gs7 = 6644.8752
	Af7 = 6644.8752
	A7  = 7040.0000
	As7 = 7458.6202
	Bf7 = 7458.6202
	B7  = 7902.1328
	C8  = 8372.0181
	Cs8 = 8869.8442
	Df8 = 8869.8442
	D8  = 9397.2726
	Ds8 = 9956.0635
	Ef8 = 9956.0635
	E8  = 10548.0818
	F8  = 11175.3034
	Fs8 = 11839.8215
	Gf8 = 11839.8215
	G8  = 12543.8540
	Gs8 = 13289.7503
	Af8 = 13289.7503
	A8  = 14080.0000
	As8 = 14917.2404
	Bf8 = 14917.2404
	B8  = 15804.2656
)
