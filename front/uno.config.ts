import {
  defineConfig,
  presetIcons,
  presetTypography,
  presetWind3,
  transformerDirectives,
  transformerVariantGroup,
} from 'unocss';

export default defineConfig({
  presets: [
    presetWind3(),
    presetTypography(),
    presetIcons({
      scale: 1.2,
      extraProperties: { display: 'inline-block', 'vertical-align': 'middle' },
    }),
  ],

  transformers: [transformerVariantGroup(), transformerDirectives()],

  content: {
    filesystem: ['src/**/*.{vue,ts,tsx,js,jsx}'],
  },

  theme: {
    colors: {
      brand: '#021726',
      'brand-hover': '#021726cc',
    },
    fontFamily: {
      sans: 'Inter, ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, sans-serif',
    },
  },

  shortcuts: {
    'ui-container': 'mx-auto w-full px-6 md:px-12',
    'ui-button':
      'inline-flex items-center justify-center gap-2 rounded-xl font-medium transition-colors cursor-pointer',
    'ui-card': 'rounded-3xl bg-white shadow-[0_0_20px_rgba(0,0,0,0.06)]',
  },
});
