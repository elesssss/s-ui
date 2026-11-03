import { createElementVNode as _createElementVNode, normalizeClass as _normalizeClass, normalizeStyle as _normalizeStyle, createVNode as _createVNode } from "vue";
// Styles
import "./VLoadingIndicator.css";

// Composables
import { useTextColor } from "../../composables/color.js";
import { makeComponentProps } from "../../composables/component.js";
import { useElementIntersection } from "../../composables/intersectionObserver.js";
import { makeSizeProps, useSize } from "../../composables/size.js";
import { makeTagProps } from "../../composables/tag.js";
import { makeThemeProps, provideTheme } from "../../composables/theme.js";
import { useToggleScope } from "../../composables/toggleScope.js"; // Utilities
import { onScopeDispose, shallowRef, toRef } from 'vue';
import { morphPath } from "./shapes.js";
import { convertToUnit, genericComponent, PREFERS_REDUCED_MOTION, propsFactory, useRender } from "../../util/index.js";
export const makeVLoadingIndicatorProps = propsFactory({
  color: String,
  contained: Boolean,
  shapeSize: [Number, String],
  ...makeComponentProps(),
  ...makeSizeProps(),
  ...makeTagProps(),
  ...makeThemeProps()
}, 'VLoadingIndicator');
export const VLoadingIndicator = genericComponent()({
  name: 'VLoadingIndicator',
  props: makeVLoadingIndicatorProps(),
  setup(props) {
    const root = shallowRef();
    const {
      themeClasses
    } = provideTheme(props);
    const {
      sizeClasses,
      sizeStyles
    } = useSize(props);
    const {
      textColorClasses,
      textColorStyles
    } = useTextColor(() => props.color);
    const {
      isIntersecting
    } = useElementIntersection(root);
    const elapsed = shallowRef(0);
    const path = toRef(() => morphPath(elapsed.value));
    useToggleScope(() => isIntersecting.value && !PREFERS_REDUCED_MOTION(), () => {
      const start = performance.now() - elapsed.value;
      let frame = requestAnimationFrame(tick);
      function tick(now) {
        // rAF timestamp is the frame start, which can precede the performance.now() taken above
        elapsed.value = Math.max(0, now - start);
        frame = requestAnimationFrame(tick);
      }
      onScopeDispose(() => cancelAnimationFrame(frame));
    });
    useRender(() => _createVNode(props.tag, {
      "ref": root,
      "class": _normalizeClass(['v-loading-indicator', {
        'v-loading-indicator--contained': props.contained
      }, themeClasses.value, sizeClasses.value, textColorClasses.value, props.class]),
      "style": _normalizeStyle([sizeStyles.value, textColorStyles.value, {
        '--v-loading-indicator-shape-size': convertToUnit(props.shapeSize)
      }, props.style]),
      "role": "progressbar"
    }, {
      default: () => [_createElementVNode("svg", {
        "class": "v-loading-indicator__shape",
        "viewBox": "-1 -1 2 2"
      }, [_createElementVNode("path", {
        "d": path.value
      }, null)])]
    }));
    return {};
  }
});
//# sourceMappingURL=VLoadingIndicator.js.map