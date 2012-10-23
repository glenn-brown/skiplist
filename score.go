// Copyright (c) 2011 Huan Du
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package skiplist

func scoreFn(key interface{}) (func(interface{}) float64) {
    switch key.(type) {
	case []byte:
		return func(key interface{}) float64 {
			t := key.([]byte)
			// only use first 8 bytes
			if len(t) > 8 {
				t = t[:8]
			}
			
			var result uint64

			for _, v := range t {
				result |= uint64(v)
				result = result << 8
			}
			return float64(result)
		}

    case float32:
		return func(t interface{}) float64 { return float64(t.(float32)) }
    case float64:
		return func(t interface{}) float64 { return float64(t.(float64)) }
    case int:
		return func(t interface{}) float64 { return float64(t.(int)) }
    case int16:
		return func(t interface{}) float64 { return float64(t.(int16)) }
    case int32:
		return func(t interface{}) float64 { return float64(t.(int32)) }
    case int64:
		return func(t interface{}) float64 { return float64(t.(int64)) }
    case int8:
		return func(t interface{}) float64 { return float64(t.(int8)) }

    case string:
		return func(key interface{}) float64 {
			t := key.(string)
			// use first 2 runes in string as score
			var runes uint64
			length := len(t)

			if length == 1 {
				runes = uint64(t[0]) << 16
			} else if length >= 2 {
				runes = uint64(t[0])<<16 + uint64(t[1])
			}
			return float64(runes)
		}

    case uint:
		return func(t interface{}) float64 { return float64(t.(uint)) }
    case uint16:
		return func(t interface{}) float64 { return float64(t.(uint16)) }
    case uint32:
		return func(t interface{}) float64 { return float64(t.(uint32)) }
    case uint64:
		return func(t interface{}) float64 { return float64(t.(uint64)) }
    case uint8:
		return func(t interface{}) float64 { return float64(t.(uint8)) }

    case uintptr:
		return func(t interface{}) float64 { return float64(t.(uintptr)) }

    case FastKey:
		return func(t interface{}) float64 { return t.(FastKey).Score() }
    }

	return func(t interface{}) float64 { return 0.0 }
}
