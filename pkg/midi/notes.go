package midi

// Note represents a note.
type Note struct {
	Name      string
	Frequency float64
	MIDI      int
}

func Names() []string {
	var result []string
	for _, note := range Notes {
		result = append(result, note.Name)
	}
	return result
}

func GetNote(name string) *Note {
	for _, note := range Notes {
		if note.Name == name {
			return &note
		}
	}
	return nil
}

var (
	C0  = Note{Name: "C0", Frequency: 16.35, MIDI: 12}
	Cs0 = Note{Name: "Cs0", Frequency: 17.32, MIDI: 13}
	Db0 = Note{Name: "Db0", Frequency: 17.32, MIDI: 13}
	D0  = Note{Name: "D0", Frequency: 18.35, MIDI: 14}
	Ds0 = Note{Name: "Ds0", Frequency: 19.45, MIDI: 15}
	Eb0 = Note{Name: "Eb0", Frequency: 19.45, MIDI: 15}
	E0  = Note{Name: "E0", Frequency: 20.60, MIDI: 16}
	F0  = Note{Name: "F0", Frequency: 21.83, MIDI: 17}
	Fs0 = Note{Name: "Fs0", Frequency: 23.12, MIDI: 18}
	Gb0 = Note{Name: "Gb0", Frequency: 23.12, MIDI: 18}
	G0  = Note{Name: "G0", Frequency: 24.50, MIDI: 19}
	Gs0 = Note{Name: "Gs0", Frequency: 25.96, MIDI: 20}
	Ab0 = Note{Name: "Ab0", Frequency: 25.96, MIDI: 20}
	A0  = Note{Name: "A0", Frequency: 27.50, MIDI: 21}
	As0 = Note{Name: "As0", Frequency: 29.14, MIDI: 22}
	Bb0 = Note{Name: "Bb0", Frequency: 29.14, MIDI: 22}
	B0  = Note{Name: "B0", Frequency: 30.87, MIDI: 23}
	C1  = Note{Name: "C1", Frequency: 32.70, MIDI: 24}
	Cs1 = Note{Name: "Cs1", Frequency: 34.65, MIDI: 25}
	Db1 = Note{Name: "Db1", Frequency: 34.65, MIDI: 25}
	D1  = Note{Name: "D1", Frequency: 36.71, MIDI: 26}
	Ds1 = Note{Name: "Ds1", Frequency: 38.89, MIDI: 27}
	Eb1 = Note{Name: "Eb1", Frequency: 38.89, MIDI: 27}
	E1  = Note{Name: "E1", Frequency: 41.20, MIDI: 28}
	F1  = Note{Name: "F1", Frequency: 43.65, MIDI: 29}
	Fs1 = Note{Name: "Fs1", Frequency: 46.25, MIDI: 30}
	Gb1 = Note{Name: "Gb1", Frequency: 46.25, MIDI: 30}
	G1  = Note{Name: "G1", Frequency: 49.00, MIDI: 31}
	Gs1 = Note{Name: "Gs1", Frequency: 51.91, MIDI: 32}
	Ab1 = Note{Name: "Ab1", Frequency: 51.91, MIDI: 32}
	A1  = Note{Name: "A1", Frequency: 55.00, MIDI: 33}
	As1 = Note{Name: "As1", Frequency: 58.27, MIDI: 34}
	Bb1 = Note{Name: "Bb1", Frequency: 58.27, MIDI: 34}
	B1  = Note{Name: "B1", Frequency: 61.74, MIDI: 35}
	C2  = Note{Name: "C2", Frequency: 65.41, MIDI: 36}
	Cs2 = Note{Name: "Cs2", Frequency: 69.30, MIDI: 37}
	Db2 = Note{Name: "Db2", Frequency: 69.30, MIDI: 37}
	D2  = Note{Name: "D2", Frequency: 73.42, MIDI: 38}
	Ds2 = Note{Name: "Ds2", Frequency: 77.78, MIDI: 39}
	Eb2 = Note{Name: "Eb2", Frequency: 77.78, MIDI: 39}
	E2  = Note{Name: "E2", Frequency: 82.41, MIDI: 40}
	F2  = Note{Name: "F2", Frequency: 87.31, MIDI: 41}
	Fs2 = Note{Name: "Fs2", Frequency: 92.50, MIDI: 42}
	Gb2 = Note{Name: "Gb2", Frequency: 92.50, MIDI: 42}
	G2  = Note{Name: "G2", Frequency: 98.00, MIDI: 43}
	Gs2 = Note{Name: "Gs2", Frequency: 103.83, MIDI: 44}
	Ab2 = Note{Name: "Ab2", Frequency: 103.83, MIDI: 44}
	A2  = Note{Name: "A2", Frequency: 110.00, MIDI: 45}
	As2 = Note{Name: "As2", Frequency: 116.54, MIDI: 46}
	Bb2 = Note{Name: "Bb2", Frequency: 116.54, MIDI: 46}
	B2  = Note{Name: "B2", Frequency: 123.47, MIDI: 47}
	C3  = Note{Name: "C3", Frequency: 130.81, MIDI: 48}
	Cs3 = Note{Name: "Cs3", Frequency: 138.59, MIDI: 49}
	Db3 = Note{Name: "Db3", Frequency: 138.59, MIDI: 49}
	D3  = Note{Name: "D3", Frequency: 146.83, MIDI: 50}
	Ds3 = Note{Name: "Ds3", Frequency: 155.56, MIDI: 51}
	Eb3 = Note{Name: "Eb3", Frequency: 155.56, MIDI: 51}
	E3  = Note{Name: "E3", Frequency: 164.81, MIDI: 52}
	F3  = Note{Name: "F3", Frequency: 174.61, MIDI: 53}
	Fs3 = Note{Name: "Fs3", Frequency: 185.00, MIDI: 54}
	Gb3 = Note{Name: "Gb3", Frequency: 185.00, MIDI: 54}
	G3  = Note{Name: "G3", Frequency: 196.00, MIDI: 55}
	Gs3 = Note{Name: "Gs3", Frequency: 207.65, MIDI: 56}
	Ab3 = Note{Name: "Ab3", Frequency: 207.65, MIDI: 56}
	A3  = Note{Name: "A3", Frequency: 220.00, MIDI: 57}
	As3 = Note{Name: "As3", Frequency: 233.08, MIDI: 58}
	Bb3 = Note{Name: "Bb3", Frequency: 233.08, MIDI: 58}
	B3  = Note{Name: "B3", Frequency: 246.94, MIDI: 59}
	C4  = Note{Name: "C4", Frequency: 261.63, MIDI: 60}
	Cs4 = Note{Name: "Cs4", Frequency: 277.18, MIDI: 61}
	Db4 = Note{Name: "Db4", Frequency: 277.18, MIDI: 61}
	D4  = Note{Name: "D4", Frequency: 293.66, MIDI: 62}
	Ds4 = Note{Name: "Ds4", Frequency: 311.13, MIDI: 63}
	Eb4 = Note{Name: "Eb4", Frequency: 311.13, MIDI: 63}
	E4  = Note{Name: "E4", Frequency: 329.63, MIDI: 64}
	F4  = Note{Name: "F4", Frequency: 349.23, MIDI: 65}
	Fs4 = Note{Name: "Fs4", Frequency: 369.99, MIDI: 66}
	Gb4 = Note{Name: "Gb4", Frequency: 369.99, MIDI: 66}
	G4  = Note{Name: "G4", Frequency: 392.00, MIDI: 67}
	Gs4 = Note{Name: "Gs4", Frequency: 415.30, MIDI: 68}
	Ab4 = Note{Name: "Ab4", Frequency: 415.30, MIDI: 68}
	A4  = Note{Name: "A4", Frequency: 440.00, MIDI: 69}
	As4 = Note{Name: "As4", Frequency: 466.16, MIDI: 70}
	Bb4 = Note{Name: "Bb4", Frequency: 466.16, MIDI: 70}
	B4  = Note{Name: "B4", Frequency: 493.88, MIDI: 71}
	C5  = Note{Name: "C5", Frequency: 523.25, MIDI: 72}
	Cs5 = Note{Name: "Cs5", Frequency: 554.37, MIDI: 73}
	Db5 = Note{Name: "Db5", Frequency: 554.37, MIDI: 73}
	D5  = Note{Name: "D5", Frequency: 587.33, MIDI: 74}
	Ds5 = Note{Name: "Ds5", Frequency: 622.25, MIDI: 75}
	Eb5 = Note{Name: "Eb5", Frequency: 622.25, MIDI: 75}
	E5  = Note{Name: "E5", Frequency: 659.25, MIDI: 76}
	F5  = Note{Name: "F5", Frequency: 698.46, MIDI: 77}
	Fs5 = Note{Name: "Fs5", Frequency: 739.99, MIDI: 78}
	Gb5 = Note{Name: "Gb5", Frequency: 739.99, MIDI: 78}
	G5  = Note{Name: "G5", Frequency: 783.99, MIDI: 79}
	Gs5 = Note{Name: "Gs5", Frequency: 830.61, MIDI: 80}
	Ab5 = Note{Name: "Ab5", Frequency: 830.61, MIDI: 80}
	A5  = Note{Name: "A5", Frequency: 880.00, MIDI: 81}
	As5 = Note{Name: "As5", Frequency: 932.33, MIDI: 82}
	Bb5 = Note{Name: "Bb5", Frequency: 932.33, MIDI: 82}
	B5  = Note{Name: "B5", Frequency: 987.77, MIDI: 83}
	C6  = Note{Name: "C6", Frequency: 1046.50, MIDI: 84}
	Cs6 = Note{Name: "Cs6", Frequency: 1108.73, MIDI: 85}
	Db6 = Note{Name: "Db6", Frequency: 1108.73, MIDI: 85}
	D6  = Note{Name: "D6", Frequency: 1174.66, MIDI: 86}
	Ds6 = Note{Name: "Ds6", Frequency: 1244.51, MIDI: 87}
	Eb6 = Note{Name: "Eb6", Frequency: 1244.51, MIDI: 87}
	E6  = Note{Name: "E6", Frequency: 1318.51, MIDI: 88}
	F6  = Note{Name: "F6", Frequency: 1396.91, MIDI: 89}
	Fs6 = Note{Name: "Fs6", Frequency: 1479.98, MIDI: 90}
	Gb6 = Note{Name: "Gb6", Frequency: 1479.98, MIDI: 90}
	G6  = Note{Name: "G6", Frequency: 1567.98, MIDI: 91}
	Gs6 = Note{Name: "Gs6", Frequency: 1661.22, MIDI: 92}
	Ab6 = Note{Name: "Ab6", Frequency: 1661.22, MIDI: 92}
	A6  = Note{Name: "A6", Frequency: 1760.00, MIDI: 93}
	As6 = Note{Name: "As6", Frequency: 1864.66, MIDI: 94}
	Bb6 = Note{Name: "Bb6", Frequency: 1864.66, MIDI: 94}
	B6  = Note{Name: "B6", Frequency: 1975.53, MIDI: 95}
	C7  = Note{Name: "C7", Frequency: 2093.00, MIDI: 96}
	Cs7 = Note{Name: "Cs7", Frequency: 2217.46, MIDI: 97}
	Db7 = Note{Name: "Db7", Frequency: 2217.46, MIDI: 97}
	D7  = Note{Name: "D7", Frequency: 2349.32, MIDI: 98}
	Ds7 = Note{Name: "Ds7", Frequency: 2489.02, MIDI: 99}
	Eb7 = Note{Name: "Eb7", Frequency: 2489.02, MIDI: 99}
	E7  = Note{Name: "E7", Frequency: 2637.02, MIDI: 100}
	F7  = Note{Name: "F7", Frequency: 2793.83, MIDI: 101}
	Fs7 = Note{Name: "Fs7", Frequency: 2959.96, MIDI: 102}
	Gb7 = Note{Name: "Gb7", Frequency: 2959.96, MIDI: 102}
	G7  = Note{Name: "G7", Frequency: 3135.96, MIDI: 103}
	Gs7 = Note{Name: "Gs7", Frequency: 3322.44, MIDI: 104}
	Ab7 = Note{Name: "Ab7", Frequency: 3322.44, MIDI: 104}
	A7  = Note{Name: "A7", Frequency: 3520.00, MIDI: 105}
	As7 = Note{Name: "As7", Frequency: 3729.31, MIDI: 106}
	Bb7 = Note{Name: "Bb7", Frequency: 3729.31, MIDI: 106}
	B7  = Note{Name: "B7", Frequency: 3951.07, MIDI: 107}
	C8  = Note{Name: "C8", Frequency: 4186.01, MIDI: 108}
	Cs8 = Note{Name: "Cs8", Frequency: 4434.92, MIDI: 109}
	Db8 = Note{Name: "Db8", Frequency: 4434.92, MIDI: 109}
	D8  = Note{Name: "D8", Frequency: 4698.63, MIDI: 110}
	Ds8 = Note{Name: "Ds8", Frequency: 4978.03, MIDI: 111}
	Eb8 = Note{Name: "Eb8", Frequency: 4978.03, MIDI: 111}
	E8  = Note{Name: "E8", Frequency: 5274.04, MIDI: 112}
	F8  = Note{Name: "F8", Frequency: 5587.65, MIDI: 113}
	Fs8 = Note{Name: "Fs8", Frequency: 5919.91, MIDI: 114}
	Gb8 = Note{Name: "Gb8", Frequency: 5919.91, MIDI: 114}
	G8  = Note{Name: "G8", Frequency: 6271.93, MIDI: 115}
	Gs8 = Note{Name: "Gs8", Frequency: 6644.88, MIDI: 116}
	Ab8 = Note{Name: "Ab8", Frequency: 6644.88, MIDI: 116}
	A8  = Note{Name: "A8", Frequency: 7040.00, MIDI: 117}
	As8 = Note{Name: "As8", Frequency: 7458.62, MIDI: 118}
	Bb8 = Note{Name: "Bb8", Frequency: 7458.62, MIDI: 118}
	B8  = Note{Name: "B8", Frequency: 7902.13, MIDI: 119}
	C9  = Note{Name: "C9", Frequency: 8372.02, MIDI: 120}
	Cs9 = Note{Name: "Cs9", Frequency: 8869.84, MIDI: 121}
	Db9 = Note{Name: "Db9", Frequency: 8869.84, MIDI: 121}
	D9  = Note{Name: "D9", Frequency: 9397.27, MIDI: 122}
	Ds9 = Note{Name: "Ds9", Frequency: 9956.06, MIDI: 123}
	Eb9 = Note{Name: "Eb9", Frequency: 9956.06, MIDI: 123}
	E9  = Note{Name: "E9", Frequency: 10548.08, MIDI: 124}
	F9  = Note{Name: "F9", Frequency: 11175.30, MIDI: 125}
	Fs9 = Note{Name: "Fs9", Frequency: 11839.82, MIDI: 126}
	Gb9 = Note{Name: "Gb9", Frequency: 11839.82, MIDI: 126}
	G9  = Note{Name: "G9", Frequency: 12543.85, MIDI: 127}
)

var (
	// Notes is a slice of all the notes in the package.
	Notes = []Note{
		C0,
		Cs0,
		Db0,
		D0,
		Ds0,
		Eb0,
		E0,
		F0,
		Fs0,
		Gb0,
		G0,
		Gs0,
		Ab0,
		A0,
		As0,
		Bb0,
		B0,
		C1,
		Cs1,
		Db1,
		D1,
		Ds1,
		Eb1,
		E1,
		F1,
		Fs1,
		Gb1,
		G1,
		Gs1,
		Ab1,
		A1,
		As1,
		Bb1,
		B1,
		C2,
		Cs2,
		Db2,
		D2,
		Ds2,
		Eb2,
		E2,
		F2,
		Fs2,
		Gb2,
		G2,
		Gs2,
		Ab2,
		A2,
		As2,
		Bb2,
		B2,
		C3,
		Cs3,
		Db3,
		D3,
		Ds3,
		Eb3,
		E3,
		F3,
		Fs3,
		Gb3,
		G3,
		Gs3,
		Ab3,
		A3,
		As3,
		Bb3,
		B3,
		C4,
		Cs4,
		Db4,
		D4,
		Ds4,
		Eb4,
		E4,
		F4,
		Fs4,
		Gb4,
		G4,
		Gs4,
		Ab4,
		A4,
		As4,
		Bb4,
		B4,
		C5,
		Cs5,
		Db5,
		D5,
		Ds5,
		Eb5,
		E5,
		F5,
		Fs5,
		Gb5,
		G5,
		Gs5,
		Ab5,
		A5,
		As5,
		Bb5,
		B5,
		C6,
		Cs6,
		Db6,
		D6,
		Ds6,
		Eb6,
		E6,
		F6,
		Fs6,
		Gb6,
		G6,
		Gs6,
		Ab6,
		A6,
		As6,
		Bb6,
		B6,
		C7,
		Cs7,
		Db7,
		D7,
		Ds7,
		Eb7,
		E7,
		F7,
		Fs7,
		Gb7,
		G7,
		Gs7,
		Ab7,
		A7,
		As7,
		Bb7,
		B7,
		C8,
		Cs8,
		Db8,
		D8,
		Ds8,
		Eb8,
		E8,
		F8,
		Fs8,
		Gb8,
		G8,
		Gs8,
		Ab8,
		A8,
		As8,
		Bb8,
		B8,
		C9,
		Cs9,
		Db9,
		D9,
		Ds9,
		Eb9,
		E9,
		F9,
		Fs9,
		Gb9,
		G9,
	}
)
