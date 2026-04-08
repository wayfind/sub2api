/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // 主色调 - 金铜色 (metask.ai style)
        primary: {
          50: '#fdf8f3',
          100: '#f9edd8',
          200: '#f0d6ae',
          300: '#dbb87a',
          400: '#c49a6c',
          500: '#b08450',
          600: '#a07845',
          700: '#8a6838',
          800: '#6d5230',
          900: '#4a3820',
          950: '#2d2010'
        },
        // 辅助色 - 暖灰
        accent: {
          50: '#f7f5f2',
          100: '#ede9e3',
          200: '#e5e2dd',
          300: '#c0bdb8',
          400: '#a8a6a0',
          500: '#6b6966',
          600: '#555550',
          700: '#3a3a3f',
          800: '#2a2928',
          900: '#1c1b1a',
          950: '#0f0e0d'
        },
        // 深色模式背景 - 暖黑
        dark: {
          50: '#f7f5f2',
          100: '#ede9e3',
          200: '#e5e2dd',
          300: '#c0bdb8',
          400: '#a8a6a0',
          500: '#6b6966',
          600: '#555550',
          700: '#2a2928',
          800: '#1c1b1a',
          900: '#141312',
          950: '#0f0e0d'
        }
      },
      fontFamily: {
        sans: [
          'Inter',
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'Helvetica Neue',
          'Arial',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'sans-serif'
        ],
        serif: [
          'Noto Serif SC',
          'Source Han Serif SC',
          'STSong',
          'SimSun',
          'Georgia',
          'serif'
        ],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      boxShadow: {
        glass: '0 8px 32px rgba(0, 0, 0, 0.12)',
        'glass-sm': '0 4px 16px rgba(0, 0, 0, 0.08)',
        glow: '0 0 20px rgba(176, 132, 80, 0.25)',
        'glow-lg': '0 0 40px rgba(197, 154, 109, 0.3), 0 4px 16px rgba(197, 154, 109, 0.25)',
        card: '0 1px 3px rgba(0, 0, 0, 0.06), 0 1px 2px rgba(0, 0, 0, 0.08)',
        'card-hover': '0 10px 40px rgba(0, 0, 0, 0.1)',
        'inner-glow': 'inset 0 1px 0 rgba(255, 255, 255, 0.1)'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-primary': 'linear-gradient(to right, #c49a6c, #a07845)',
        'gradient-dark': 'linear-gradient(135deg, #1c1b1a 0%, #0f0e0d 100%)',
        'gradient-glass':
          'linear-gradient(135deg, rgba(255,255,255,0.1) 0%, rgba(255,255,255,0.05) 100%)',
        'mesh-gradient':
          'radial-gradient(at 40% 20%, rgba(176, 132, 80, 0.06) 0px, transparent 50%), radial-gradient(at 80% 0%, rgba(197, 154, 109, 0.04) 0px, transparent 50%), radial-gradient(at 0% 50%, rgba(176, 132, 80, 0.04) 0px, transparent 50%)'
      },
      animation: {
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        shimmer: 'shimmer 2s linear infinite',
        glow: 'glow 2s ease-in-out infinite alternate'
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(20px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' }
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.95)' },
          '100%': { opacity: '1', transform: 'scale(1)' }
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' }
        },
        glow: {
          '0%': { boxShadow: '0 0 20px rgba(176, 132, 80, 0.25)' },
          '100%': { boxShadow: '0 0 30px rgba(176, 132, 80, 0.4)' }
        }
      },
      backdropBlur: {
        xs: '2px'
      },
      borderRadius: {
        '4xl': '2rem'
      }
    }
  },
  plugins: []
}
