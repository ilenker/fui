package calc

import (
	"math"
	"time"
	"unsafe"
)

// temporary import of my math "library"
/*······················································································kVec2    */
type Vec2 struct {
	X int
	Y int
}

func (v1 Vec2) AddNoWrap(v2 Vec2) Vec2 {
	return Vec2{
		X: v1.X + v2.X,
		Y: v1.Y + v2.Y,
	}
}

func (v1 Vec2) Translate(angleRad float64, distance float64) Vec2 {
	dx := distance * math.Cos(angleRad)
	dy := distance * math.Sin(angleRad)

	newX := float64(v1.X) + dx
	newY := float64(v1.Y) + dy

	newVec := Vec2{int(math.Round(newX)), int(math.Round(newY))}

	return newVec
}

func Dist(v1, v2 Vec2) float64 {
	return math.Sqrt(
		math.Pow(float64(v1.X-v2.X), 2) +
			math.Pow(float64(v1.Y-v2.Y), 2),
	)
}

/*······················································································kVec3    */
type Vec3[T int | int8 | int16 | int32 | float32 | float64] struct {
	x T
	y T
	z T
}

type Slice3f64 struct {
	rs []float64
	gs []float64
	bs []float64
}

func CopyRGB(destSlice *Slice3f64, index int, r, g, b float64) {
	destSlice.rs[index] = r
	destSlice.gs[index] = g
	destSlice.bs[index] = b
}

func (v1 *Vec3[float32]) Scale(v2 Vec3[float32]) Vec3[float32] {
	return Vec3[float32]{
		v1.x * v2.x,
		v1.y * v2.y,
		v1.z * v2.z,
	}
}

/*······················································································kVecRGB  */
type VecRGB struct {
	r int32
	g int32
	b int32
}

func (v1 VecRGB) Add(v2 VecRGB) VecRGB {

	v3 := v1
	v3.r = v1.r + v2.r
	v3.b = v1.b + v2.b
	v3.g = v1.g + v2.g

	if v3.r > 255 {
		v3.r = 255
	}
	if v3.g > 255 {
		v3.g = 255
	}
	if v3.b > 255 {
		v3.b = 255
	}

	if v3.r < 0 {
		v3.r = 0
	}
	if v3.g < 0 {
		v3.g = 0
	}
	if v3.b < 0 {
		v3.b = 0
	}

	return v3
}

func NewVecRGB[T int | int8 | int16 | int32 | int64](r T, g T, b T) VecRGB {
	return VecRGB{int32(r), int32(g), int32(b)}
}

/*······················································································kMath    */
func AbsInt(n int) int {
	if n < 0 {
		return ^n + 1
	}
	return n
}

func AbsInt16(n int16) int16 {
	if n < 0 {
		return ^n + 1
	}
	return n
}

func Lerp(x, y int, f float64) float64 {
	return float64(x)*(1.0-f) + float64(y)*f
}

func Lerp64(x, y, f float64) float64 {
	return x*(1.0-f) + (y * f)
}

func ILerp(x, y int, num float64) float64 {
	if x == y {
		return 0
	}
	return (num - float64(x)) / float64(y-x)
}

func ILerp32(x, y int, num float32) float32 {
	if x == y {
		return 0
	}
	return (num - float32(x)) / float32(y-x)
}

func ILerp64[T int | float64](x, y T, num float64) float64 {
	if x == y {
		return 0
	}
	return (num - float64(x)) / float64(y-x)
}

func WrapInt(n, n_max int) int {
	wrap := ((n % n_max) + n_max) % n_max
	return wrap
}

func MakeAverageDurationBuffer(size int) func(time.Duration) (int64, []time.Duration) {
	buffer := make([]time.Duration, size)
	i := 0

	return func(d time.Duration) (int64, []time.Duration) {
		buffer[i] = d
		i = WrapInt(i+1, size)
		total := time.Duration(0)

		for n := range size {
			total += buffer[n]
		}

		return (total.Microseconds() / int64(size)), buffer
	}
}

func MakeAverageIntBuffer(size int) func(n int) (float64, []int) {
	buffer := make([]int, size)
	i := 0
	return func(n int) (float64, []int) {
		buffer[i] = n
		i = WrapInt(i+1, size)
		total := 0
		for n := range size {
			total += buffer[n]
		}
		return (float64(total) / float64(size)), buffer
	}
}

func Clamp[T int | int8 | int16 | int32 | float32 | float64](n, min, max T) T {
	if n > max {
		return max
	}
	if n < min {
		return min
	}
	return n
}

func ClampMin[T int | int8 | int16 | int32 | float32 | float64](n, min T) T {
	if n < min {
		return min
	}
	return n
}

func B2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// https://dev.to/chigbeef_77/bool-int-but-stupid-in-go-3jb3
// Fast boolean to integer
func FB2i(b bool) int {
	return int(*(*byte)(unsafe.Pointer(&b)))
}

func FloatEq(a, b, ε float64) bool {
	diff := math.Abs(a - b)
	if diff <= ε {
		return true
	}
	return diff <= ε*math.Max(math.Abs(a), math.Abs(b))
}

// https://stackoverflow.com/a/2074403
//   - Assume int32
func Abs(n int) int {
	value := int32(n)
	temp := value >> 31 // make a mask of the sign bit
	value ^= temp       // toggle the bits if value is negative
	value += temp & 1   // add one if value was negative
	return int(value)
}

func Last[T any](s []T) (T, bool) {
	var value T
	if len(s) == 0 {
		return value, false
	}
	return s[len(s)-1], true
}
