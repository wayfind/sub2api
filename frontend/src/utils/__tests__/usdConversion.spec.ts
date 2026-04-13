import { describe, expect, it } from 'vitest'
import { USD_TO_U, formatUsdFromU, uToUsdRound, usdToURound } from '../format'

describe('USD ↔ U 数值转换', () => {
  describe('uToUsdRound', () => {
    it('null / undefined / NaN 返回 null', () => {
      expect(uToUsdRound(null)).toBeNull()
      expect(uToUsdRound(undefined)).toBeNull()
      expect(uToUsdRound(NaN)).toBeNull()
    })

    it('0 返回 0', () => {
      expect(uToUsdRound(0)).toBe(0)
    })

    it('整数 U 转 USD', () => {
      expect(uToUsdRound(70)).toBe(1)
      expect(uToUsdRound(700)).toBe(10)
      expect(uToUsdRound(70000)).toBe(1000)
    })

    it('非整除 U 转 USD 时 round 到 4 位小数', () => {
      // 25000000 / 70 = 357142.857142857... → ×10000 = 3571428571.4285... → round = 3571428571 → /10000
      expect(uToUsdRound(25000000)).toBe(357142.8571)
      // 800000 / 70 = 11428.5714285714... → ×10000 = 114285714.2857... → round = 114285714 → /10000
      expect(uToUsdRound(800000)).toBe(11428.5714)
      // 1 / 70 = 0.0142857142... → ×10000 = 142.857... → round = 143 → /10000
      expect(uToUsdRound(1)).toBe(0.0143)
    })

    it('小金额', () => {
      // 0.7 / 70 = 0.01 → ×10000 = 100 → round = 100 → /10000
      expect(uToUsdRound(0.7)).toBe(0.01)
      // 0.07 / 70 = 0.001 → ×10000 = 10 → round = 10 → /10000
      expect(uToUsdRound(0.07)).toBe(0.001)
    })
  })

  describe('usdToURound', () => {
    it('null / undefined / NaN 返回 null', () => {
      expect(usdToURound(null)).toBeNull()
      expect(usdToURound(undefined)).toBeNull()
      expect(usdToURound(NaN)).toBeNull()
    })

    it('0 返回 0', () => {
      expect(usdToURound(0)).toBe(0)
    })

    it('整数 USD 转 U', () => {
      expect(usdToURound(1)).toBe(70)
      expect(usdToURound(10)).toBe(700)
      expect(usdToURound(1000)).toBe(70000)
    })

    it('两位小数 USD 转 U', () => {
      expect(usdToURound(2.5)).toBe(175)
      expect(usdToURound(0.01)).toBe(0.7)
      expect(usdToURound(0.05)).toBe(3.5)
    })
  })

  describe('round-trip 一致性', () => {
    // 这是关键不变量：用户输入一个 USD 值，存到 DB（U），下次加载又显示成 USD，
    // 在合理的 USD 输入精度（≤ 2 位小数）下应该恒等。
    const usdInputs = [
      0, 0.01, 0.05, 0.1, 0.5, 1, 2.5, 10, 100, 142.86, 1000, 11428.57, 357142.86,
    ]

    for (const usd of usdInputs) {
      it(`USD=${usd} round-trip 恒等`, () => {
        const u = usdToURound(usd)
        expect(u).not.toBeNull()
        const usdBack = uToUsdRound(u!)
        expect(usdBack).toBe(usd)
      })
    }

    it('null 在 round-trip 中保持 null', () => {
      expect(uToUsdRound(usdToURound(null))).toBeNull()
      expect(usdToURound(uToUsdRound(null))).toBeNull()
    })

    // 反向：从 DB 已有的 U 值出发，加载 → 编辑（不变）→ 保存
    // U 值从生产 DB 实际观察到的样本
    const uInputs = [
      0, 100, 1000, 10000, 80000, 800000, 25000000,
    ]

    for (const u of uInputs) {
      it(`U=${u} 不变更编辑后保存恒等`, () => {
        // 不应该因为 round 而漂移
        const usd = uToUsdRound(u)
        expect(usd).not.toBeNull()
        const uBack = usdToURound(usd)
        // 漂移上界推导：
        //   uToUsdRound 内部做 round(u/70 × 10000)/10000，最大舍入误差 ±0.00005 USD
        //   usdToURound 把这个 USD 值 ×70 还原 → ±0.00005 × 70 = ±0.0035 U
        //   第二步本身又会引入 ±0.00005 USD ≈ ±0.0035/10000 U，相对前者可忽略
        // 因此 round-trip 漂移上界严格 ≤ 0.0035 U
        const drift = Math.abs((uBack ?? 0) - u)
        expect(drift).toBeLessThanOrEqual(0.0035)
      })
    }

    // 上界紧致性测试：确认 0.0035 这个上界不是过于宽松，至少有一个真实样本能逼近它。
    // 这是上面 uInputs 里"看似没毛病的整数 U 样本"覆盖不到的——它们除以 70 之后
    // 大部分小数会落在 round 边界的安全区域，drift 远小于 0.0035。
    // 如果未来有人误以为"实际 drift 极小"而把上界缩小到 0.001，这个测试会立刻挂。
    it('round-trip 漂移在最坏情况下能逼近 0.0035 U 上界', () => {
      // 找一个 u 使得 u/70 的 ×10000 后小数部分 ≈ 0.5（半位舍入边界）。
      // u/70 × 10000 = n + 0.5  ⇒  u = (n + 0.5) × 0.007
      // 取 n=5000 → u = 35.0035。
      // 验算：35.0035 / 70 = 0.500050 → ×10000 = 5000.5
      // Math.round(5000.5) → JS 是 half-away-from-zero，但浮点表示 5000.5
      //   实际浮点值可能略小于 5000.5，导致 round 到 5000；这取决于 IEEE 754 表示。
      // 不论 round 到 5000 还是 5001，drift 都 ≈ 0.0035 U。
      const u = 35.0035
      const uBack = usdToURound(uToUsdRound(u))
      const drift = Math.abs((uBack ?? 0) - u)
      expect(drift).toBeLessThanOrEqual(0.0035)
      // 紧致性：drift 应该明显大于 0.001（上界至少不能被收紧 3 倍以上）
      expect(drift).toBeGreaterThan(0.001)
    })

    // 另一个最坏情况样本：使用更大的整数避免浮点表示精度干扰
    it('大数最坏情况 round-trip 也在上界内', () => {
      // u = 35000.0035 → /70 = 500.0000500 → ×10000 = 5000000.5
      // round 到 5000000 或 5000001 → /10000 = 500.0000 或 500.0001
      // ×70 = 35000.0 或 35000.007 → drift ≈ 0.0035
      const u = 35000.0035
      const uBack = usdToURound(uToUsdRound(u))
      const drift = Math.abs((uBack ?? 0) - u)
      expect(drift).toBeLessThanOrEqual(0.0035)
      expect(drift).toBeGreaterThan(0.001)
    })
  })

  describe('formatUsdFromU 边界值', () => {
    it('0 → "$0.000000"', () => {
      // 0 < 0.01 → 6 位小数分支
      expect(formatUsdFromU(0)).toBe('$0.000000')
    })

    it('微小金额走 6 位小数分支', () => {
      // 0.07 / 70 = 0.001 → < 0.01
      expect(formatUsdFromU(0.07)).toBe('$0.001000')
    })

    it('中等金额走 4 位小数分支', () => {
      // 35 / 70 = 0.5 → >= 0.01 且 < 1
      expect(formatUsdFromU(35)).toBe('$0.5000')
      // 0.7 / 70 = 0.01 → 命中 >= 0.01 边界
      expect(formatUsdFromU(0.7)).toBe('$0.0100')
    })

    it('大金额走 2 位小数分支', () => {
      // 70 / 70 = 1.0 → 命中 >= 1 边界
      expect(formatUsdFromU(70)).toBe('$1.00')
      // 25000000 / 70 ≈ 357142.857
      expect(formatUsdFromU(25000000)).toBe('$357142.86')
    })

    it('USD_TO_U 常数 = 70（重构守卫）', () => {
      // 如果未来汇率改了这个测试要更新——同时所有 helper 也需要重新审计
      expect(USD_TO_U).toBe(70)
    })
  })
})
