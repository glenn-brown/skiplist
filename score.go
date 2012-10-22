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

type Scorer interface {
	Score() float64
}

func score(key interface{}) (score float64) {
    switch t := key.(type) {
    case []byte:
        // only use first 8 bytes
        if len(t) > 8 {
            t = t[:8]
        }

        var result uint64

        for _, v := range t {
            result |= uint64(v)
            result = result << 8
        }

        score = float64(result)

    case float32:
        score = float64(t)

    case float64:
        score = t

    case int:
        score = float64(t)

    case int16:
        score = float64(t)

    case int32:
        score = float64(t)

    case int64:
        score = float64(t)

    case int8:
        score = float64(t)

    case string:
        // use first 2 runes in string as score
        var runes uint64
        length := len(t)

        if length == 1 {
            runes = uint64(t[0]) << 16
        } else if length >= 2 {
            runes = uint64(t[0])<<16 + uint64(t[1])
        }

        score = float64(runes)

    case uint:
        score = float64(t)

    case uint16:
        score = float64(t)

    case uint32:
        score = float64(t)

    case uint64:
        score = float64(t)

    case uint8:
        score = float64(t)

    case uintptr:
        score = float64(t)

    case Scorer:
        score = t.Score()
    }

    return
}
