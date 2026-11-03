import { normalizeClass as _normalizeClass, normalizeStyle as _normalizeStyle, createElementVNode as _createElementVNode, Fragment as _Fragment, createVNode as _createVNode } from "vue";
// Styles
import "./VProgressLinear.css";

// Composables
import { useBackgroundColor, useTextColor } from "../../composables/color.js";
import { makeComponentProps } from "../../composables/component.js";
import { useElementIntersection } from "../../composables/intersectionObserver.js";
import { useRtl } from "../../composables/locale.js";
import { makeLocationProps, useLocation } from "../../composables/location.js";
import { useProxiedModel } from "../../composables/proxiedModel.js";
import { useResizeObserver } from "../../composables/resizeObserver.js";
import { makeRevealProps, useReveal } from "../../composables/reveal.js";
import { makeRoundedProps, useRounded } from "../../composables/rounded.js";
import { makeTagProps } from "../../composables/tag.js";
import { makeThemeProps, provideTheme } from "../../composables/theme.js";
import { useToggleScope } from "../../composables/toggleScope.js"; // Utilities
import { computed, shallowRef, toRef, Transition } from 'vue';
import { makeChunksProps, useChunks } from "./chunks.js";
import { useWavy, WAVE_AMPLITUDE } from "./wavy.js";
import { clamp, convertToUnit, genericComponent, isNullOrUndefined, isObject, propsFactory, useRender } from "../../util/index.js"; // Types
export const makeVProgressLinearProps = propsFactory({
  absolute: Boolean,
  active: {
    type: Boolean,
    default: true
  },
  bgColor: String,
  bgOpacity: [Number, String],
  bufferValue: {
    type: [Number, String],
    default: 0
  },
  bufferColor: String,
  bufferOpacity: [Number, String],
  clickable: Boolean,
  color: String,
  height: {
    type: [Number, String],
    default: 4
  },
  indeterminate: Boolean,
  max: {
    type: [Number, String],
    default: 100
  },
  modelValue: {
    type: [Number, String],
    default: 0
  },
  opacity: [Number, String],
  reverse: Boolean,
  shape: {
    type: String,
    default: 'flat',
    validator: v => ['flat', 'wavy'].includes(v)
  },
  stopIndicator: {
    type: Boolean,
    default: undefined
  },
  stream: Boolean,
  striped: Boolean,
  roundedBar: Boolean,
  transition: {
    type: [Boolean, Object],
    default: undefined
  },
  ...makeChunksProps(),
  ...makeComponentProps(),
  ...makeLocationProps({
    location: 'top'
  }),
  ...makeRevealProps(),
  ...makeRoundedProps(),
  ...makeTagProps(),
  ...makeThemeProps()
}, 'VProgressLinear');
export const VProgressLinear = genericComponent()({
  name: 'VProgressLinear',
  props: makeVProgressLinearProps(),
  emits: {
    'update:modelValue': value => true
  },
  setup(props, {
    slots
  }) {
    const root = shallowRef();
    const progress = useProxiedModel(props, 'modelValue');
    const {
      isRtl,
      rtlClasses
    } = useRtl();
    const {
      themeClasses
    } = provideTheme(props);
    const {
      locationStyles
    } = useLocation(props);
    const {
      textColorClasses,
      textColorStyles
    } = useTextColor(() => props.color);
    const {
      backgroundColorClasses,
      backgroundColorStyles
    } = useBackgroundColor(() => props.bgColor || props.color);
    const {
      backgroundColorClasses: bufferColorClasses,
      backgroundColorStyles: bufferColorStyles
    } = useBackgroundColor(() => props.bufferColor || props.bgColor || props.color);
    const {
      backgroundColorClasses: barColorClasses,
      backgroundColorStyles: barColorStyles
    } = useBackgroundColor(() => props.color);
    const {
      textColorClasses: trackColorClasses,
      textColorStyles: trackColorStyles
    } = useTextColor(() => props.bgColor);
    const {
      roundedClasses,
      roundedStyles
    } = useRounded(props);
    const {
      isIntersecting
    } = useElementIntersection(root);
    const {
      state: revealState,
      duration: revealDuration
    } = useReveal(props);
    const max = computed(() => parseFloat(props.max));
    const height = computed(() => convertToUnit(props.height));
    const normalizedBuffer = computed(() => clamp(parseFloat(props.bufferValue) / max.value * 100, 0, 100));
    const normalizedValue = computed(() => revealState.value === 'initial' ? 0 : clamp(parseFloat(progress.value) / max.value * 100, 0, 100));
    const isReversed = computed(() => isRtl.value !== props.reverse);
    const transitionDuration = computed(() => props.transition === false ? undefined : convertToUnit(isObject(props.transition) ? props.transition.duration : undefined, 'ms'));
    const bgOpacity = toRef(() => isNullOrUndefined(props.bgOpacity) ? undefined : parseFloat(props.bgOpacity));
    const isWavy = toRef(() => props.shape === 'wavy');
    // undefined keeps split without the stop until the next major
    const hasStopIndicator = toRef(() => !props.indeterminate && (isWavy.value ? props.stopIndicator !== false : props.variant === 'split' && props.stopIndicator === true));
    const transitionName = computed(() => props.indeterminate ? 'fade-transition' : 'slide-x-transition');
    const containerWidth = shallowRef(0);
    const {
      hasChunks,
      splitStyles,
      chunksMaskStyles,
      snapValueToChunk
    } = useChunks(props, containerWidth, normalizedValue, normalizedBuffer, isReversed);
    useToggleScope(() => hasChunks.value || isWavy.value, () => {
      useResizeObserver(root, entries => {
        containerWidth.value = entries[0].contentRect.width;
      });
    });
    const bufferWidth = computed(() => {
      return hasChunks.value ? snapValueToChunk(normalizedBuffer.value) : normalizedBuffer.value;
    });
    const barWidth = computed(() => {
      return hasChunks.value ? snapValueToChunk(normalizedValue.value) : normalizedValue.value;
    });
    const {
      geometry,
      svgEl,
      waveEl
    } = useWavy(props, {
      containerWidth,
      value: normalizedValue,
      isActive: () => props.active && isIntersecting.value,
      duration: () => {
        if (revealState.value === 'pending') return revealDuration.value;
        const duration = getComputedStyle(root.value).getPropertyValue('--v-progress-linear-transition-duration').trim();
        return parseFloat(duration) * (duration.endsWith('ms') ? 1 : 1000);
      }
    });
    function handleClick(e) {
      if (!root.value) return;
      const {
        left,
        right,
        width
      } = root.value.getBoundingClientRect();
      const value = isReversed.value ? width - e.clientX + (right - width) : e.clientX - left;
      progress.value = Math.round(value / width * max.value);
    }
    function renderBackgroundBar() {
      return _createElementVNode("div", {
        "class": _normalizeClass(['v-progress-linear__background', backgroundColorClasses.value]),
        "style": _normalizeStyle([backgroundColorStyles.value, {
          opacity: bgOpacity.value,
          width: props.stream ? 0 : undefined
        }, props.indeterminate ? {} : splitStyles.value?.background])
      }, null);
    }
    function renderWavy() {
      if (!geometry.value) return null;
      const {
        width,
        stroke,
        wavelength,
        center,
        end,
        gap,
        track,
        wave,
        offset
      } = geometry.value;
      function renderTrack() {
        return _createElementVNode("path", {
          "class": _normalizeClass(['v-progress-linear__track', trackColorClasses.value]),
          "style": _normalizeStyle([trackColorStyles.value, {
            opacity: bgOpacity.value
          }]),
          "d": track,
          "pathLength": "100"
        }, null);
      }
      function renderWave(bar) {
        return _createElementVNode("path", {
          "class": _normalizeClass(['v-progress-linear__wave', bar]),
          "d": wave,
          "pathLength": 100 + offset
        }, null);
      }
      return _createElementVNode("svg", {
        "ref": svgEl,
        "class": _normalizeClass(['v-progress-linear__wavy', {
          'v-progress-linear__wavy--indeterminate': props.indeterminate
        }, textColorClasses.value]),
        "style": _normalizeStyle([textColorStyles.value, {
          '--v-progress-linear-stroke': convertToUnit(stroke),
          '--v-progress-linear-wavelength': convertToUnit(wavelength),
          '--v-progress-linear-wave-offset': offset,
          '--v-progress-linear-wave-gap': gap
        }]),
        "viewBox": `0 0 ${width} ${center * 2}`
      }, [props.indeterminate ? _createElementVNode(_Fragment, null, [renderTrack(), renderTrack(), renderTrack(), renderWave('long'), renderWave('short')]) : _createElementVNode(_Fragment, null, [renderTrack(), hasStopIndicator.value && _createElementVNode("circle", {
        "class": "v-progress-linear__stop",
        "cx": end,
        "cy": center
      }, null), _createElementVNode("path", {
        "ref": waveEl,
        "class": "v-progress-linear__wave"
      }, null)])]);
    }
    useRender(() => _createVNode(props.tag, {
      "ref": root,
      "class": _normalizeClass(['v-progress-linear', {
        'v-progress-linear--absolute': props.absolute,
        'v-progress-linear--active': props.active && isIntersecting.value,
        'v-progress-linear--reverse': isReversed.value,
        'v-progress-linear--rounded': props.rounded,
        'v-progress-linear--rounded-bar': props.roundedBar,
        'v-progress-linear--striped': props.striped,
        'v-progress-linear--clickable': props.clickable,
        'v-progress-linear--no-transition': props.transition === false,
        'v-progress-linear--revealing': ['initial', 'pending'].includes(revealState.value),
        'v-progress-linear--variant-split': props.variant === 'split',
        'v-progress-linear--shape-wavy': isWavy.value
      }, roundedClasses.value, themeClasses.value, rtlClasses.value, props.class]),
      "style": _normalizeStyle([{
        bottom: props.location === 'bottom' ? 0 : undefined,
        top: props.location === 'top' ? 0 : undefined,
        height: !props.active ? 0 : isWavy.value ? convertToUnit(parseFloat(props.height) + 2 * WAVE_AMPLITUDE) : height.value,
        '--v-progress-linear-height': height.value,
        '--v-progress-linear-transition-duration': transitionDuration.value,
        '--v-progress-reveal-duration': `${revealDuration.value}ms`,
        '--v-progress-chunk-gap': convertToUnit(props.chunkGap),
        ...(props.absolute ? locationStyles.value : {})
      }, isWavy.value ? undefined : chunksMaskStyles.value, roundedStyles.value, props.style]),
      "role": "progressbar",
      "aria-hidden": props.active ? 'false' : 'true',
      "aria-valuemin": "0",
      "aria-valuemax": props.max,
      "aria-valuenow": props.indeterminate ? undefined : Math.min(parseFloat(progress.value), max.value),
      "onClick": props.clickable && handleClick
    }, {
      default: () => [isWavy.value ? renderWavy() : _createElementVNode(_Fragment, null, [props.stream && _createElementVNode("div", {
        "key": "stream",
        "class": _normalizeClass(['v-progress-linear__stream', textColorClasses.value]),
        "style": {
          ...textColorStyles.value,
          [isReversed.value ? 'left' : 'right']: `calc(${height.value} * -1)`,
          borderTop: `calc(${height.value} / 2) dotted`,
          opacity: props.bufferOpacity != null ? parseFloat(props.bufferOpacity) : undefined,
          top: `calc(50% - ${height.value} / 4)`,
          width: convertToUnit(100 - normalizedBuffer.value, '%'),
          '--v-progress-linear-stream-to': `calc(${height.value} * ${isReversed.value ? 1 : -1})`
        }
      }, null), (props.variant !== 'split' || !props.indeterminate) && renderBackgroundBar(), _createElementVNode("div", {
        "class": _normalizeClass(['v-progress-linear__buffer', bufferColorClasses.value]),
        "style": _normalizeStyle([bufferColorStyles.value, {
          opacity: props.bufferOpacity != null ? parseFloat(props.bufferOpacity) : undefined,
          width: convertToUnit(bufferWidth.value, '%')
        }, splitStyles.value?.buffer])
      }, null), hasStopIndicator.value && _createElementVNode("div", {
        "class": _normalizeClass(['v-progress-linear__stop', barColorClasses.value]),
        "style": _normalizeStyle(barColorStyles.value)
      }, null), _createVNode(Transition, {
        "name": transitionName.value
      }, {
        default: () => [!props.indeterminate ? _createElementVNode("div", {
          "class": _normalizeClass(['v-progress-linear__determinate', barColorClasses.value]),
          "style": _normalizeStyle([barColorStyles.value, {
            width: convertToUnit(barWidth.value, '%')
          }, splitStyles.value?.bar])
        }, null) : _createElementVNode("div", {
          "class": "v-progress-linear__indeterminate"
        }, [props.variant === 'split' && _createElementVNode(_Fragment, null, [renderBackgroundBar(), renderBackgroundBar(), renderBackgroundBar()]), ['long', 'short'].map(bar => _createElementVNode("div", {
          "key": bar,
          "class": _normalizeClass(['v-progress-linear__indeterminate', bar, barColorClasses.value]),
          "style": _normalizeStyle(barColorStyles.value)
        }, null))])]
      })]), slots.default && _createElementVNode("div", {
        "class": "v-progress-linear__content"
      }, [slots.default({
        value: normalizedValue.value,
        buffer: normalizedBuffer.value
      })])]
    }));
    return {};
  }
});
//# sourceMappingURL=VProgressLinear.js.map