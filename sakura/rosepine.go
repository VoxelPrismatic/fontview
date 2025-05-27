package sakura

type HSLVector struct {
	H float64
	L float64
	S float64
}

type SakuraPaint[T any] struct {
	Love T // Red
	Gold T // Yellow
	Rose T // Pink
	Pine T // Darker blue
	Foam T // Light blue
	Iris T // Vibrant blue
	Tree T // Green
}

type SakuraHl[T any] struct {
	High T
	Med  T
	Low  T
}

type SakuraLayer[T any] struct {
	Base    T
	Overlay T
	Surface T
	Inverse T
	None    T
}

type SakuraText[T any] struct {
	Normal T
	Muted  T
	Subtle T
}

type SakuraPalette[T any] struct {
	Paint SakuraPaint[T]
	Hl    SakuraHl[T]
	Layer SakuraLayer[T]
	Text  SakuraText[T]
}

type SakuraSwatch[T any] struct {
	Dawn SakuraPalette[T]
	Moon SakuraPalette[T]
	Main SakuraPalette[T]
}

type VectorTheme[T any] struct {
	Base T // Background color
	Text T // Text color
}

type VectorPalette[T any] struct {
	Paint SakuraPaint[T]
	Dawn  VectorTheme[T]
	Moon  VectorTheme[T]
	Main  VectorTheme[T]
}

type DeriveVector[T any] struct {
	SakuraSwatch[T]
	Paint SakuraPaint[T]
}

func MapSwatch[T, R any](swatch SakuraSwatch[T], cb func(T) R) SakuraSwatch[R] {
	return SakuraSwatch[R]{
		Dawn: MapPalette(swatch.Dawn, cb),
		Moon: MapPalette(swatch.Moon, cb),
		Main: MapPalette(swatch.Main, cb),
	}
}

func MapPalette[T, R any](p SakuraPalette[T], cb func(T) R) SakuraPalette[R] {
	return SakuraPalette[R]{
		Paint: MapPaint(p.Paint, cb),
		Hl:    MapHl(p.Hl, cb),
		Layer: MapLayer(p.Layer, cb),
		Text:  MapText(p.Text, cb),
	}
}

func MapPaint[T, R any](p SakuraPaint[T], cb func(T) R) SakuraPaint[R] {
	return SakuraPaint[R]{
		Love: cb(p.Love),
		Rose: cb(p.Rose),
		Gold: cb(p.Gold),
		Iris: cb(p.Iris),
		Tree: cb(p.Tree),
		Foam: cb(p.Foam),
		Pine: cb(p.Pine),
	}
}

func MapHl[T, R any](p SakuraHl[T], cb func(T) R) SakuraHl[R] {
	return SakuraHl[R]{
		High: cb(p.High),
		Med:  cb(p.Med),
		Low:  cb(p.Low),
	}
}

func MapLayer[T, R any](p SakuraLayer[T], cb func(T) R) SakuraLayer[R] {
	return SakuraLayer[R]{
		Base:    cb(p.Base),
		Overlay: cb(p.Overlay),
		Surface: cb(p.Surface),
	}
}

func MapText[T, R any](p SakuraText[T], cb func(T) R) SakuraText[R] {
	return SakuraText[R]{
		Normal: cb(p.Normal),
		Muted:  cb(p.Muted),
		Subtle: cb(p.Subtle),
	}
}
