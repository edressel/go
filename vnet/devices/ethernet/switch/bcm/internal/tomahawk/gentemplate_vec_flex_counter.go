// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// autogenerated: do not edit!
// generated from gentemplate [gentemplate -d Package=tomahawk -id flex_counter -d VecType=flex_counter_vec -d Type=flex_counter github.com/platinasystems/go/elib/vec.tmpl]

package tomahawk

import (
	"github.com/platinasystems/go/elib"
)

type flex_counter_vec []flex_counter

func (p *flex_counter_vec) Resize(n uint) {
	c := elib.Index(cap(*p))
	l := elib.Index(len(*p)) + elib.Index(n)
	if l > c {
		c = elib.NextResizeCap(l)
		q := make([]flex_counter, l, c)
		copy(q, *p)
		*p = q
	}
	*p = (*p)[:l]
}

func (p *flex_counter_vec) validate(new_len uint, zero *flex_counter) *flex_counter {
	c := elib.Index(cap(*p))
	lʹ := elib.Index(len(*p))
	l := elib.Index(new_len)
	if l <= c {
		// Need to reslice to larger length?
		if l >= lʹ {
			*p = (*p)[:l]
		}
		return &(*p)[l-1]
	}
	return p.validateSlowPath(zero, c, l, lʹ)
}

func (p *flex_counter_vec) validateSlowPath(zero *flex_counter,
	c, l, lʹ elib.Index) *flex_counter {
	if l > c {
		cNext := elib.NextResizeCap(l)
		q := make([]flex_counter, cNext, cNext)
		copy(q, *p)
		if zero != nil {
			for i := c; i < cNext; i++ {
				q[i] = *zero
			}
		}
		*p = q[:l]
	}
	if l > lʹ {
		*p = (*p)[:l]
	}
	return &(*p)[l-1]
}

func (p *flex_counter_vec) Validate(i uint) *flex_counter {
	return p.validate(i+1, (*flex_counter)(nil))
}

func (p *flex_counter_vec) ValidateInit(i uint, zero flex_counter) *flex_counter {
	return p.validate(i+1, &zero)
}

func (p *flex_counter_vec) ValidateLen(l uint) (v *flex_counter) {
	if l > 0 {
		v = p.validate(l, (*flex_counter)(nil))
	}
	return
}

func (p *flex_counter_vec) ValidateLenInit(l uint, zero flex_counter) (v *flex_counter) {
	if l > 0 {
		v = p.validate(l, &zero)
	}
	return
}

func (p flex_counter_vec) Len() uint { return uint(len(p)) }