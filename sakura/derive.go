package sakura

import (
	"cmp"
	"math"
)

type RGB struct {
	R int
	G int
	B int
}

func clamp[T cmp.Ordered](min_ T, val T, max_ T) T {
	return max(min_, min(val, max_))
}

func (rgb RGB) FromHexInt(hex int) RGB {
	return RGB{
		R: (hex & 0xff0000) >> 16,
		G: (hex & 0x00ff00) >> 8,
		B: (hex & 0x0000ff),
	}
}

func (rgb RGB) ToHexInt() int {
	return (rgb.R << 16) + (rgb.G << 8) + rgb.B
}

func (rgb RGB) Hsl() HSLVector {
	r := float64(rgb.R) / 255
	g := float64(rgb.G) / 255
	b := float64(rgb.B) / 255

	t_min := min(r, g, b)
	t_max := max(r, g, b)

	ret := HSLVector{
		H: 0,
		S: (t_max - t_min) / t_max,
		L: t_max,
	}
	t_diff := t_min - t_max

	if t_diff == 0 {
		// pass
	} else if r == t_max {
		ret.H = (g - b) / t_diff
		if ret.H < 0 {
			ret.H += 6
		}
	} else if g == t_max {
		ret.H = (b-r)/t_diff + 2
	} else {
		ret.H = (r-g)/t_diff + 4
	}

	for ret.H > 360 {
		ret.H -= 360
	}
	for ret.H < 0 {
		ret.H += 360
	}

	return ret
}

func (hsl HSLVector) Rgb() RGB {
	h := hsl.H / 360
	s := clamp(0, hsl.S/100, 1)
	v := clamp(0, hsl.L/100, 1)

	var r, g, b float64

	int_part := math.Floor(h * 6)
	float_part := h*6 - int_part

	p := v * (1 - s)
	q := v * (1 - float_part*s)
	t := v * (1 - (1-float_part)*s)

	switch int_part {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	case 5:
		r, g, b = v, p, q
	}

	return RGB{
		R: int(math.Floor(r*255 + 0.5)),
		G: int(math.Floor(g*255 + 0.5)),
		B: int(math.Floor(b*255 + 0.5)),
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

func (v VectorPalette[int]) Parse() SakuraSwatch[int] {
	ret := SakuraSwatch[int]{}
	ret.Dawn.Paint =

}
