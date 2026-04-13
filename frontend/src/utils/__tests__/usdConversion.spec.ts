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
      // 25000000 / 70 = 357142.857142857...
      expect(uToUsdRound(25000000)).toBe(357142.8571)
      // 800000 / 70 = 11428.571428...
      expect(uToUsdRound(800000)).toBe(11428.5714)
      // 1 / 70 = 0.01428571... → round 到 4 位 = 0.0143
      expect(uToUsdRound(1)).toBe(0.0143)
    })

    it('小金额', () => {
      // 0.7 / 70 = 0.01
      expect(uToUsdRound(0.7)).toBe(0.01)
      // 0.07 / 70 = 0.001
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
        // 注意：当 u/70 不是 4 位小数能精确表达时，会有微小漂移
        // 这是预期的 lossy 转换，4 位小数下漂移 ≤ 70 * 5e-5 = 0.0035 U
        const drift = Math.abs((uBack ?? 0) - u)
        expect(drift).toBeLessThanOrEqual(0.0035)
      })
    }
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
