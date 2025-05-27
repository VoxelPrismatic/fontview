package sakura

import (
	"cmp"
	"encoding/json"
	"fmt"
	"math"
)

type RGB struct {
	R int
	G int
	B int
}

func clamp[T cmp.Ordered](min_, val, max_ T) T {
	return min(max(min_, val), max_)
}

func (rgb RGB) FromHexInt(hex int) RGB {
	return RGB{
		R: (hex & 0xff0000) >> 16,
		G: (hex & 0x00ff00) >> 8,
		B: (hex & 0x0000ff),
	}
}

func (rgb RGB) ToHexInt() int {
	return (clamp(0, rgb.R, 255) << 16) + (clamp(0, rgb.G, 255) << 8) + clamp(0, rgb.B, 255)
}

var _int = map[bool]int{
	true:  1,
	false: 0,
}

func (rgb RGB) Hsl() HSLVector {
	r := float64(rgb.R) / 255
	g := float64(rgb.G) / 255
	b := float64(rgb.B) / 255

	v_min := min(r, g, b)
	v_max := max(r, g, b)
	delta := v_max - v_min

	ret := HSLVector{
		H: 0,
		S: (v_max - v_min) / v_max,
		L: v_max,
	}

	if delta == 0 {
		// no color component
	} else if r == v_max {
		ret.H = (g - b) / delta
	} else if g == v_max {
		ret.H = (b-r)/delta + 2
	} else {
		ret.H = (r-g)/delta + 4
	}

	for ret.H > 6 {
		ret.H -= 6
	}
	for ret.H < 0 {
		ret.H += 6
	}

	ret.H *= 60
	ret.S *= 100
	ret.L *= 100

	return ret
}

func (hsl HSLVector) Rgb() RGB {
	hue := hsl.H / 360
	sat := clamp(0, hsl.S/100, 1)
	lum := clamp(0, hsl.L/100, 1)

	var r, g, b float64

	int_part := math.Floor(hue * 6)
	float_part := hue*6 - int_part

	pilot := lum * (1 - sat)
	quart := lum * (1 - float_part*sat)
	third := lum * (1 - (1-float_part)*sat)

	switch int_part {
	case 0:
		r, g, b = lum, third, pilot
	case 1:
		r, g, b = quart, lum, pilot
	case 2:
		r, g, b = pilot, lum, third
	case 3:
		r, g, b = pilot, quart, lum
	case 4:
		r, g, b = third, pilot, lum
	default:
		r, g, b = lum, pilot, quart
	}

	return RGB{
		R: clamp(0, int(math.Round(r*255)), 255),
		G: clamp(0, int(math.Round(g*255)), 255),
		B: clamp(0, int(math.Round(b*255)), 255),
	}
}

func (vec HSLVector) Calc(source, target int) HSLVector {
	src := RGB{}.FromHexInt(source).Hsl()
	trg := RGB{}.FromHexInt(target).Hsl()

	return HSLVector{
		H: trg.H - src.H,
		S: trg.S - src.S,
		L: trg.L - src.L,
	}
}

func (vec HSLVector) Tx(hex int) int {
	rgb := RGB{}.FromHexInt(hex)
	hsl := rgb.Hsl()
	tx := HSLVector{
		H: hsl.H + vec.H,
		S: hsl.S + vec.S,
		L: hsl.L + vec.L,
	}
	rgb = tx.Rgb()
	return rgb.ToHexInt()
}

func (v DerivePalette) Parse() SakuraSwatch[int] {
	ret := SakuraSwatch[int]{}

	ret.Dawn.Paint = v.Paint
	ret.Moon.Paint = MergePaint(v.Paint, Vectors.Moon.Paint)
	ret.Main.Paint = MergePaint(v.Paint, Vectors.Main.Paint)

	ret.Dawn.Hl = DeriveHl(v.Dawn.Base, Vectors.Dawn.Hl)
	ret.Moon.Hl = DeriveHl(v.Moon.Base, Vectors.Moon.Hl)
	ret.Main.Hl = DeriveHl(v.Main.Base, Vectors.Main.Hl)

	ret.Dawn.Layer = DeriveLayer(v.Dawn.Base, Vectors.Dawn.Layer)
	ret.Moon.Layer = DeriveLayer(v.Moon.Base, Vectors.Moon.Layer)
	ret.Main.Layer = DeriveLayer(v.Main.Base, Vectors.Main.Layer)

	ret.Dawn.Text = DeriveText(v.Dawn.Text, Vectors.Dawn.Text)
	ret.Moon.Text = DeriveText(v.Moon.Text, Vectors.Moon.Text)
	ret.Main.Text = DeriveText(v.Main.Text, Vectors.Main.Text)

	return ret
}

func Test() {
	b, err := json.Marshal(MapSwatch(
		Sakura.Parse(),
		func(c int) string { return fmt.Sprintf("#%06x", c) },
	))
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
