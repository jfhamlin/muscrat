package ugen

// void LeakDC_next_i_4(LeakDC* unit, int inNumSamples) {
//     float* out = ZOUT(0);
//     float* in = ZIN(0);
//     double b1 = unit->m_b1;
//     double y1 = unit->m_y1;
//     double x1 = unit->m_x1;

//     LOOP1(inNumSamples / 4, double x00 = ZXP(in); double x01 = ZXP(in); double x02 = ZXP(in); double x03 = ZXP(in);
//           float out0 = y1 = x00 - x1 + b1 * y1; float out1 = y1 = x01 - x00 + b1 * y1;
//           float out2 = y1 = x02 - x01 + b1 * y1; float out3 = y1 = x03 - x02 + b1 * y1;

//           ZXP(out) = out0; ZXP(out) = out1; ZXP(out) = out2; ZXP(out) = out3;

//           x1 = x03;);
//     unit->m_x1 = x1;
//     unit->m_y1 = zapgremlins(y1);
// }

// void LeakDC_next_i(LeakDC* unit, int inNumSamples) {
//     float* out = ZOUT(0);
//     float* in = ZIN(0);
//     double b1 = unit->m_b1;
//     double y1 = unit->m_y1;
//     double x1 = unit->m_x1;

//     LOOP1(inNumSamples, double x0 = ZXP(in); ZXP(out) = y1 = x0 - x1 + b1 * y1; x1 = x0;);
//     unit->m_x1 = x1;
//     unit->m_y1 = zapgremlins(y1);
// }

// void LeakDC_next(LeakDC* unit, int inNumSamples) {
//     if (ZIN0(1) == unit->m_b1) {
//         if ((inNumSamples & 3) == 0)
//             LeakDC_next_i_4(unit, inNumSamples);
//         else
//             LeakDC_next_i(unit, inNumSamples);
//     } else {
//         float* out = ZOUT(0);
//         float* in = ZIN(0);
//         double b1 = unit->m_b1;
//         unit->m_b1 = ZIN0(1);

//         double y1 = unit->m_y1;
//         double x1 = unit->m_x1;

//         double b1_slope = CALCSLOPE(unit->m_b1, b1);
//         LOOP1(inNumSamples, double x0 = ZXP(in); ZXP(out) = y1 = x0 - x1 + b1 * y1; x1 = x0; b1 += b1_slope;);
//         unit->m_x1 = x1;
//         unit->m_y1 = zapgremlins(y1);
//     }
// }

// void LeakDC_next_1(LeakDC* unit, int inNumSamples) {
//     double b1 = unit->m_b1 = ZIN0(1);

//     double y1 = unit->m_y1;
//     double x1 = unit->m_x1;

//     double x0 = ZIN0(0);
//     ZOUT0(0) = y1 = x0 - x1 + b1 * y1;
//     x1 = x0;

//     unit->m_x1 = x1;
//     unit->m_y1 = zapgremlins(y1);
// }

// void LeakDC_Ctor(LeakDC* unit) {
//     // printf("LeakDC_Ctor\n");
//     if (BUFLENGTH == 1)
//         SETCALC(LeakDC_next_1);
//     else {
//         if (INRATE(1) == calc_ScalarRate) {
//             if ((BUFLENGTH & 3) == 0)
//                 SETCALC(LeakDC_next_i_4);
//             else
//                 SETCALC(LeakDC_next_i);
//         } else
//             SETCALC(LeakDC_next);
//     }
//     unit->m_b1 = 0.0;
//     unit->m_x1 = ZIN0(0);
//     unit->m_y1 = 0.0;
//     LeakDC_next_1(unit, 1);
// }

type LeakDC struct {
	b1 float64
	x1 float64
	y1 float64

	initialized bool
}

// func NewLeakDC() UGen {
// 	return &LeakDC{}
// }

// func (l *LeakDC) Gen(ctx context.Context, cfg SampleConfig, out []float64) {
// 	in := cfg.InputSamples["in"]
// 	coef := cfg.InputSamples["coef"]

// 	_ = in[len(out)-1]
// 	_ = coef[len(out)-1]

// 	if !l.initialized {
// 		l.x1 = in[0]
// 		out[0] = leakDCNext1(in, b1)
// 		l.initialized = true
// 	}
// }

// func (l *LeakDC) leakDCNext1(in []float64, coef []float64) float64 {
// 	l.b1 = coef[0]
// 	b1 := l.b1

// 	y1 := l.y1
// 	x1 := l.x1

// 	x0 := in[0]
// 	y1 := x0 - x1 + b1*y1
// 	x1 = x0

// 	l.x1 = x1
// 	l.y1 = y1

// 	return y1
// }
