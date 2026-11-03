const require_meta = require("./meta.js");
const require_processor = require("./processor.js");
const require_array_bracket_newline = require("./rules/array-bracket-newline.js");
const require_array_bracket_spacing = require("./rules/array-bracket-spacing.js");
const require_array_element_newline = require("./rules/array-element-newline.js");
const require_arrow_spacing = require("./rules/arrow-spacing.js");
const require_attribute_hyphenation = require("./rules/attribute-hyphenation.js");
const require_attributes_order = require("./rules/attributes-order.js");
const require_block_lang = require("./rules/block-lang.js");
const require_block_order = require("./rules/block-order.js");
const require_block_spacing = require("./rules/block-spacing.js");
const require_block_tag_newline = require("./rules/block-tag-newline.js");
const require_brace_style = require("./rules/brace-style.js");
const require_camelcase = require("./rules/camelcase.js");
const require_comma_dangle = require("./rules/comma-dangle.js");
const require_comma_spacing = require("./rules/comma-spacing.js");
const require_comma_style = require("./rules/comma-style.js");
const require_comment_directive = require("./rules/comment-directive.js");
const require_component_api_style = require("./rules/component-api-style.js");
const require_component_definition_name_casing = require("./rules/component-definition-name-casing.js");
const require_component_name_in_template_casing = require("./rules/component-name-in-template-casing.js");
const require_component_options_name_casing = require("./rules/component-options-name-casing.js");
const require_custom_event_name_casing = require("./rules/custom-event-name-casing.js");
const require_define_emits_declaration = require("./rules/define-emits-declaration.js");
const require_define_macros_order = require("./rules/define-macros-order.js");
const require_define_props_declaration = require("./rules/define-props-declaration.js");
const require_define_props_destructuring = require("./rules/define-props-destructuring.js");
const require_dot_location = require("./rules/dot-location.js");
const require_dot_notation = require("./rules/dot-notation.js");
const require_enforce_style_attribute = require("./rules/enforce-style-attribute.js");
const require_eqeqeq = require("./rules/eqeqeq.js");
const require_first_attribute_linebreak = require("./rules/first-attribute-linebreak.js");
const require_func_call_spacing = require("./rules/func-call-spacing.js");
const require_html_button_has_type = require("./rules/html-button-has-type.js");
const require_html_closing_bracket_newline = require("./rules/html-closing-bracket-newline.js");
const require_html_closing_bracket_spacing = require("./rules/html-closing-bracket-spacing.js");
const require_html_comment_content_newline = require("./rules/html-comment-content-newline.js");
const require_html_comment_content_spacing = require("./rules/html-comment-content-spacing.js");
const require_html_comment_indent = require("./rules/html-comment-indent.js");
const require_html_end_tags = require("./rules/html-end-tags.js");
const require_html_indent = require("./rules/html-indent.js");
const require_html_quotes = require("./rules/html-quotes.js");
const require_html_self_closing = require("./rules/html-self-closing.js");
const require_jsx_uses_vars = require("./rules/jsx-uses-vars.js");
const require_key_spacing = require("./rules/key-spacing.js");
const require_keyword_spacing = require("./rules/keyword-spacing.js");
const require_match_component_file_name = require("./rules/match-component-file-name.js");
const require_match_component_import_name = require("./rules/match-component-import-name.js");
const require_max_attributes_per_line = require("./rules/max-attributes-per-line.js");
const require_max_len = require("./rules/max-len.js");
const require_max_lines_per_block = require("./rules/max-lines-per-block.js");
const require_max_props = require("./rules/max-props.js");
const require_max_template_depth = require("./rules/max-template-depth.js");
const require_multi_word_component_names = require("./rules/multi-word-component-names.js");
const require_multiline_html_element_content_newline = require("./rules/multiline-html-element-content-newline.js");
const require_multiline_ternary = require("./rules/multiline-ternary.js");
const require_mustache_interpolation_spacing = require("./rules/mustache-interpolation-spacing.js");
const require_new_line_between_multi_line_property = require("./rules/new-line-between-multi-line-property.js");
const require_next_tick_style = require("./rules/next-tick-style.js");
const require_no_arrow_functions_in_watch = require("./rules/no-arrow-functions-in-watch.js");
const require_no_async_in_computed_properties = require("./rules/no-async-in-computed-properties.js");
const require_no_bare_strings_in_template = require("./rules/no-bare-strings-in-template.js");
const require_no_boolean_default = require("./rules/no-boolean-default.js");
const require_no_child_content = require("./rules/no-child-content.js");
const require_no_computed_properties_in_data = require("./rules/no-computed-properties-in-data.js");
const require_no_console = require("./rules/no-console.js");
const require_no_constant_condition = require("./rules/no-constant-condition.js");
const require_no_content_outside_block = require("./rules/no-content-outside-block.js");
const require_no_custom_modifiers_on_v_model = require("./rules/no-custom-modifiers-on-v-model.js");
const require_no_define_model_default = require("./rules/no-define-model-default.js");
const require_no_deprecated_data_object_declaration = require("./rules/no-deprecated-data-object-declaration.js");
const require_no_deprecated_delete_set = require("./rules/no-deprecated-delete-set.js");
const require_no_deprecated_destroyed_lifecycle = require("./rules/no-deprecated-destroyed-lifecycle.js");
const require_no_deprecated_dollar_listeners_api = require("./rules/no-deprecated-dollar-listeners-api.js");
const require_no_deprecated_dollar_scopedslots_api = require("./rules/no-deprecated-dollar-scopedslots-api.js");
const require_no_deprecated_events_api = require("./rules/no-deprecated-events-api.js");
const require_no_deprecated_filter = require("./rules/no-deprecated-filter.js");
const require_no_deprecated_functional_template = require("./rules/no-deprecated-functional-template.js");
const require_no_deprecated_html_element_is = require("./rules/no-deprecated-html-element-is.js");
const require_no_deprecated_inline_template = require("./rules/no-deprecated-inline-template.js");
const require_no_deprecated_model_definition = require("./rules/no-deprecated-model-definition.js");
const require_no_deprecated_props_default_this = require("./rules/no-deprecated-props-default-this.js");
const require_no_deprecated_router_link_tag_prop = require("./rules/no-deprecated-router-link-tag-prop.js");
const require_no_deprecated_scope_attribute = require("./rules/no-deprecated-scope-attribute.js");
const require_no_deprecated_slot_attribute = require("./rules/no-deprecated-slot-attribute.js");
const require_no_deprecated_slot_scope_attribute = require("./rules/no-deprecated-slot-scope-attribute.js");
const require_no_deprecated_v_bind_sync = require("./rules/no-deprecated-v-bind-sync.js");
const require_no_deprecated_v_is = require("./rules/no-deprecated-v-is.js");
const require_no_deprecated_v_on_native_modifier = require("./rules/no-deprecated-v-on-native-modifier.js");
const require_no_deprecated_v_on_number_modifiers = require("./rules/no-deprecated-v-on-number-modifiers.js");
const require_no_deprecated_vue_config_keycodes = require("./rules/no-deprecated-vue-config-keycodes.js");
const require_no_dupe_keys = require("./rules/no-dupe-keys.js");
const require_no_dupe_v_else_if = require("./rules/no-dupe-v-else-if.js");
const require_no_duplicate_attr_inheritance = require("./rules/no-duplicate-attr-inheritance.js");
const require_no_duplicate_attributes = require("./rules/no-duplicate-attributes.js");
const require_no_duplicate_class_names = require("./rules/no-duplicate-class-names.js");
const require_no_empty_component_block = require("./rules/no-empty-component-block.js");
const require_no_empty_pattern = require("./rules/no-empty-pattern.js");
const require_no_export_in_script_setup = require("./rules/no-export-in-script-setup.js");
const require_no_expose_after_await = require("./rules/no-expose-after-await.js");
const require_no_extra_parens = require("./rules/no-extra-parens.js");
const require_no_implicit_coercion = require("./rules/no-implicit-coercion.js");
const require_no_import_compiler_macros = require("./rules/no-import-compiler-macros.js");
const require_no_irregular_whitespace = require("./rules/no-irregular-whitespace.js");
const require_no_lifecycle_after_await = require("./rules/no-lifecycle-after-await.js");
const require_no_literals_in_template = require("./rules/no-literals-in-template.js");
const require_no_lone_template = require("./rules/no-lone-template.js");
const require_no_loss_of_precision = require("./rules/no-loss-of-precision.js");
const require_no_multi_spaces = require("./rules/no-multi-spaces.js");
const require_no_multiple_objects_in_class = require("./rules/no-multiple-objects-in-class.js");
const require_no_multiple_slot_args = require("./rules/no-multiple-slot-args.js");
const require_no_multiple_template_root = require("./rules/no-multiple-template-root.js");
const require_no_mutating_props = require("./rules/no-mutating-props.js");
const require_no_negated_condition = require("./rules/no-negated-condition.js");
const require_no_negated_v_if_condition = require("./rules/no-negated-v-if-condition.js");
const require_no_parsing_error = require("./rules/no-parsing-error.js");
const require_no_potential_component_option_typo = require("./rules/no-potential-component-option-typo.js");
const require_no_props_shadow = require("./rules/no-props-shadow.js");
const require_no_ref_as_operand = require("./rules/no-ref-as-operand.js");
const require_no_ref_object_reactivity_loss = require("./rules/no-ref-object-reactivity-loss.js");
const require_no_required_prop_with_default = require("./rules/no-required-prop-with-default.js");
const require_no_reserved_component_names = require("./rules/no-reserved-component-names.js");
const require_no_reserved_keys = require("./rules/no-reserved-keys.js");
const require_no_reserved_props = require("./rules/no-reserved-props.js");
const require_no_restricted_block = require("./rules/no-restricted-block.js");
const require_no_restricted_call_after_await = require("./rules/no-restricted-call-after-await.js");
const require_no_restricted_class = require("./rules/no-restricted-class.js");
const require_no_restricted_component_names = require("./rules/no-restricted-component-names.js");
const require_no_restricted_component_options = require("./rules/no-restricted-component-options.js");
const require_no_restricted_custom_event = require("./rules/no-restricted-custom-event.js");
const require_no_restricted_html_elements = require("./rules/no-restricted-html-elements.js");
const require_no_restricted_props = require("./rules/no-restricted-props.js");
const require_no_restricted_static_attribute = require("./rules/no-restricted-static-attribute.js");
const require_no_restricted_syntax = require("./rules/no-restricted-syntax.js");
const require_no_restricted_v_bind = require("./rules/no-restricted-v-bind.js");
const require_no_restricted_v_on = require("./rules/no-restricted-v-on.js");
const require_no_root_v_if = require("./rules/no-root-v-if.js");
const require_no_setup_props_reactivity_loss = require("./rules/no-setup-props-reactivity-loss.js");
const require_no_shadow_native_events = require("./rules/no-shadow-native-events.js");
const require_no_shared_component_data = require("./rules/no-shared-component-data.js");
const require_no_side_effects_in_computed_properties = require("./rules/no-side-effects-in-computed-properties.js");
const require_no_spaces_around_equal_signs_in_attribute = require("./rules/no-spaces-around-equal-signs-in-attribute.js");
const require_no_sparse_arrays = require("./rules/no-sparse-arrays.js");
const require_no_static_inline_styles = require("./rules/no-static-inline-styles.js");
const require_no_template_key = require("./rules/no-template-key.js");
const require_no_template_shadow = require("./rules/no-template-shadow.js");
const require_no_template_target_blank = require("./rules/no-template-target-blank.js");
const require_no_textarea_mustache = require("./rules/no-textarea-mustache.js");
const require_no_this_in_before_route_enter = require("./rules/no-this-in-before-route-enter.js");
const require_no_undef_components = require("./rules/no-undef-components.js");
const require_no_undef_directives = require("./rules/no-undef-directives.js");
const require_no_undef_properties = require("./rules/no-undef-properties.js");
const require_no_unsupported_features = require("./rules/no-unsupported-features.js");
const require_no_unused_components = require("./rules/no-unused-components.js");
const require_no_unused_emit_declarations = require("./rules/no-unused-emit-declarations.js");
const require_no_unused_properties = require("./rules/no-unused-properties.js");
const require_no_unused_refs = require("./rules/no-unused-refs.js");
const require_no_unused_vars = require("./rules/no-unused-vars.js");
const require_no_use_computed_property_like_method = require("./rules/no-use-computed-property-like-method.js");
const require_no_use_v_else_with_v_for = require("./rules/no-use-v-else-with-v-for.js");
const require_no_use_v_if_with_v_for = require("./rules/no-use-v-if-with-v-for.js");
const require_no_useless_concat = require("./rules/no-useless-concat.js");
const require_no_useless_mustaches = require("./rules/no-useless-mustaches.js");
const require_no_useless_template_attributes = require("./rules/no-useless-template-attributes.js");
const require_no_useless_v_bind = require("./rules/no-useless-v-bind.js");
const require_no_v_for_template_key_on_child = require("./rules/no-v-for-template-key-on-child.js");
const require_no_v_for_template_key = require("./rules/no-v-for-template-key.js");
const require_no_v_html = require("./rules/no-v-html.js");
const require_no_v_model_argument = require("./rules/no-v-model-argument.js");
const require_no_v_text_v_html_on_component = require("./rules/no-v-text-v-html-on-component.js");
const require_no_v_text = require("./rules/no-v-text.js");
const require_no_watch_after_await = require("./rules/no-watch-after-await.js");
const require_object_curly_newline = require("./rules/object-curly-newline.js");
const require_object_curly_spacing = require("./rules/object-curly-spacing.js");
const require_object_property_newline = require("./rules/object-property-newline.js");
const require_object_shorthand = require("./rules/object-shorthand.js");
const require_one_component_per_file = require("./rules/one-component-per-file.js");
const require_operator_linebreak = require("./rules/operator-linebreak.js");
const require_order_in_components = require("./rules/order-in-components.js");
const require_padding_line_between_blocks = require("./rules/padding-line-between-blocks.js");
const require_padding_line_between_tags = require("./rules/padding-line-between-tags.js");
const require_padding_lines_in_component_definition = require("./rules/padding-lines-in-component-definition.js");
const require_prefer_define_options = require("./rules/prefer-define-options.js");
const require_prefer_import_from_vue = require("./rules/prefer-import-from-vue.js");
const require_prefer_prop_type_boolean_first = require("./rules/prefer-prop-type-boolean-first.js");
const require_prefer_separate_static_class = require("./rules/prefer-separate-static-class.js");
const require_prefer_single_event_payload = require("./rules/prefer-single-event-payload.js");
const require_prefer_template = require("./rules/prefer-template.js");
const require_prefer_true_attribute_shorthand = require("./rules/prefer-true-attribute-shorthand.js");
const require_prefer_use_template_ref = require("./rules/prefer-use-template-ref.js");
const require_prefer_v_model = require("./rules/prefer-v-model.js");
const require_prop_name_casing = require("./rules/prop-name-casing.js");
const require_quote_props = require("./rules/quote-props.js");
const require_require_component_is = require("./rules/require-component-is.js");
const require_require_default_export = require("./rules/require-default-export.js");
const require_require_default_prop = require("./rules/require-default-prop.js");
const require_require_direct_export = require("./rules/require-direct-export.js");
const require_require_emit_validator = require("./rules/require-emit-validator.js");
const require_require_explicit_emits = require("./rules/require-explicit-emits.js");
const require_require_explicit_slots = require("./rules/require-explicit-slots.js");
const require_require_expose = require("./rules/require-expose.js");
const require_require_macro_variable_name = require("./rules/require-macro-variable-name.js");
const require_require_name_property = require("./rules/require-name-property.js");
const require_require_prop_comment = require("./rules/require-prop-comment.js");
const require_require_prop_type_constructor = require("./rules/require-prop-type-constructor.js");
const require_require_prop_types = require("./rules/require-prop-types.js");
const require_require_render_return = require("./rules/require-render-return.js");
const require_require_slots_as_functions = require("./rules/require-slots-as-functions.js");
const require_require_toggle_inside_transition = require("./rules/require-toggle-inside-transition.js");
const require_require_typed_object_prop = require("./rules/require-typed-object-prop.js");
const require_require_typed_ref = require("./rules/require-typed-ref.js");
const require_require_v_for_key = require("./rules/require-v-for-key.js");
const require_require_valid_default_prop = require("./rules/require-valid-default-prop.js");
const require_restricted_component_names = require("./rules/restricted-component-names.js");
const require_return_in_computed_property = require("./rules/return-in-computed-property.js");
const require_return_in_emits_validator = require("./rules/return-in-emits-validator.js");
const require_script_indent = require("./rules/script-indent.js");
const require_singleline_html_element_content_newline = require("./rules/singleline-html-element-content-newline.js");
const require_slot_name_casing = require("./rules/slot-name-casing.js");
const require_sort_keys = require("./rules/sort-keys.js");
const require_space_in_parens = require("./rules/space-in-parens.js");
const require_space_infix_ops = require("./rules/space-infix-ops.js");
const require_space_unary_ops = require("./rules/space-unary-ops.js");
const require_static_class_names_order = require("./rules/static-class-names-order.js");
const require_template_curly_spacing = require("./rules/template-curly-spacing.js");
const require_this_in_template = require("./rules/this-in-template.js");
const require_use_v_on_exact = require("./rules/use-v-on-exact.js");
const require_v_bind_style = require("./rules/v-bind-style.js");
const require_v_for_delimiter_style = require("./rules/v-for-delimiter-style.js");
const require_v_if_else_key = require("./rules/v-if-else-key.js");
const require_v_on_event_hyphenation = require("./rules/v-on-event-hyphenation.js");
const require_v_on_handler_style = require("./rules/v-on-handler-style.js");
const require_v_on_style = require("./rules/v-on-style.js");
const require_v_slot_style = require("./rules/v-slot-style.js");
const require_valid_attribute_name = require("./rules/valid-attribute-name.js");
const require_valid_define_emits = require("./rules/valid-define-emits.js");
const require_valid_define_options = require("./rules/valid-define-options.js");
const require_valid_define_props = require("./rules/valid-define-props.js");
const require_valid_model_definition = require("./rules/valid-model-definition.js");
const require_valid_next_tick = require("./rules/valid-next-tick.js");
const require_valid_template_root = require("./rules/valid-template-root.js");
const require_valid_v_bind_sync = require("./rules/valid-v-bind-sync.js");
const require_valid_v_bind = require("./rules/valid-v-bind.js");
const require_valid_v_cloak = require("./rules/valid-v-cloak.js");
const require_valid_v_else_if = require("./rules/valid-v-else-if.js");
const require_valid_v_else = require("./rules/valid-v-else.js");
const require_valid_v_for = require("./rules/valid-v-for.js");
const require_valid_v_html = require("./rules/valid-v-html.js");
const require_valid_v_if = require("./rules/valid-v-if.js");
const require_valid_v_is = require("./rules/valid-v-is.js");
const require_valid_v_memo = require("./rules/valid-v-memo.js");
const require_valid_v_model = require("./rules/valid-v-model.js");
const require_valid_v_on = require("./rules/valid-v-on.js");
const require_valid_v_once = require("./rules/valid-v-once.js");
const require_valid_v_pre = require("./rules/valid-v-pre.js");
const require_valid_v_show = require("./rules/valid-v-show.js");
const require_valid_v_slot = require("./rules/valid-v-slot.js");
const require_valid_v_text = require("./rules/valid-v-text.js");
// lib/plugin.ts
var plugin_default = {
	meta: require_meta.default,
	rules: {
		"array-bracket-newline": require_array_bracket_newline.default,
		"array-bracket-spacing": require_array_bracket_spacing.default,
		"array-element-newline": require_array_element_newline.default,
		"arrow-spacing": require_arrow_spacing.default,
		"attribute-hyphenation": require_attribute_hyphenation.default,
		"attributes-order": require_attributes_order.default,
		"block-lang": require_block_lang.default,
		"block-order": require_block_order.default,
		"block-spacing": require_block_spacing.default,
		"block-tag-newline": require_block_tag_newline.default,
		"brace-style": require_brace_style.default,
		camelcase: require_camelcase.default,
		"comma-dangle": require_comma_dangle.default,
		"comma-spacing": require_comma_spacing.default,
		"comma-style": require_comma_style.default,
		"comment-directive": require_comment_directive.default,
		"component-api-style": require_component_api_style.default,
		"component-definition-name-casing": require_component_definition_name_casing.default,
		"component-name-in-template-casing": require_component_name_in_template_casing.default,
		"component-options-name-casing": require_component_options_name_casing.default,
		"custom-event-name-casing": require_custom_event_name_casing.default,
		"define-emits-declaration": require_define_emits_declaration.default,
		"define-macros-order": require_define_macros_order.default,
		"define-props-declaration": require_define_props_declaration.default,
		"define-props-destructuring": require_define_props_destructuring.default,
		"dot-location": require_dot_location.default,
		"dot-notation": require_dot_notation.default,
		"enforce-style-attribute": require_enforce_style_attribute.default,
		eqeqeq: require_eqeqeq.default,
		"first-attribute-linebreak": require_first_attribute_linebreak.default,
		"func-call-spacing": require_func_call_spacing.default,
		"html-button-has-type": require_html_button_has_type.default,
		"html-closing-bracket-newline": require_html_closing_bracket_newline.default,
		"html-closing-bracket-spacing": require_html_closing_bracket_spacing.default,
		"html-comment-content-newline": require_html_comment_content_newline.default,
		"html-comment-content-spacing": require_html_comment_content_spacing.default,
		"html-comment-indent": require_html_comment_indent.default,
		"html-end-tags": require_html_end_tags.default,
		"html-indent": require_html_indent.default,
		"html-quotes": require_html_quotes.default,
		"html-self-closing": require_html_self_closing.default,
		"jsx-uses-vars": require_jsx_uses_vars.default,
		"key-spacing": require_key_spacing.default,
		"keyword-spacing": require_keyword_spacing.default,
		"match-component-file-name": require_match_component_file_name.default,
		"match-component-import-name": require_match_component_import_name.default,
		"max-attributes-per-line": require_max_attributes_per_line.default,
		"max-len": require_max_len.default,
		"max-lines-per-block": require_max_lines_per_block.default,
		"max-props": require_max_props.default,
		"max-template-depth": require_max_template_depth.default,
		"multi-word-component-names": require_multi_word_component_names.default,
		"multiline-html-element-content-newline": require_multiline_html_element_content_newline.default,
		"multiline-ternary": require_multiline_ternary.default,
		"mustache-interpolation-spacing": require_mustache_interpolation_spacing.default,
		"new-line-between-multi-line-property": require_new_line_between_multi_line_property.default,
		"next-tick-style": require_next_tick_style.default,
		"no-arrow-functions-in-watch": require_no_arrow_functions_in_watch.default,
		"no-async-in-computed-properties": require_no_async_in_computed_properties.default,
		"no-bare-strings-in-template": require_no_bare_strings_in_template.default,
		"no-boolean-default": require_no_boolean_default.default,
		"no-child-content": require_no_child_content.default,
		"no-computed-properties-in-data": require_no_computed_properties_in_data.default,
		"no-console": require_no_console.default,
		"no-constant-condition": require_no_constant_condition.default,
		"no-content-outside-block": require_no_content_outside_block.default,
		"no-custom-modifiers-on-v-model": require_no_custom_modifiers_on_v_model.default,
		"no-define-model-default": require_no_define_model_default.default,
		"no-deprecated-data-object-declaration": require_no_deprecated_data_object_declaration.default,
		"no-deprecated-delete-set": require_no_deprecated_delete_set.default,
		"no-deprecated-destroyed-lifecycle": require_no_deprecated_destroyed_lifecycle.default,
		"no-deprecated-dollar-listeners-api": require_no_deprecated_dollar_listeners_api.default,
		"no-deprecated-dollar-scopedslots-api": require_no_deprecated_dollar_scopedslots_api.default,
		"no-deprecated-events-api": require_no_deprecated_events_api.default,
		"no-deprecated-filter": require_no_deprecated_filter.default,
		"no-deprecated-functional-template": require_no_deprecated_functional_template.default,
		"no-deprecated-html-element-is": require_no_deprecated_html_element_is.default,
		"no-deprecated-inline-template": require_no_deprecated_inline_template.default,
		"no-deprecated-model-definition": require_no_deprecated_model_definition.default,
		"no-deprecated-props-default-this": require_no_deprecated_props_default_this.default,
		"no-deprecated-router-link-tag-prop": require_no_deprecated_router_link_tag_prop.default,
		"no-deprecated-scope-attribute": require_no_deprecated_scope_attribute.default,
		"no-deprecated-slot-attribute": require_no_deprecated_slot_attribute.default,
		"no-deprecated-slot-scope-attribute": require_no_deprecated_slot_scope_attribute.default,
		"no-deprecated-v-bind-sync": require_no_deprecated_v_bind_sync.default,
		"no-deprecated-v-is": require_no_deprecated_v_is.default,
		"no-deprecated-v-on-native-modifier": require_no_deprecated_v_on_native_modifier.default,
		"no-deprecated-v-on-number-modifiers": require_no_deprecated_v_on_number_modifiers.default,
		"no-deprecated-vue-config-keycodes": require_no_deprecated_vue_config_keycodes.default,
		"no-dupe-keys": require_no_dupe_keys.default,
		"no-dupe-v-else-if": require_no_dupe_v_else_if.default,
		"no-duplicate-attr-inheritance": require_no_duplicate_attr_inheritance.default,
		"no-duplicate-attributes": require_no_duplicate_attributes.default,
		"no-duplicate-class-names": require_no_duplicate_class_names.default,
		"no-empty-component-block": require_no_empty_component_block.default,
		"no-empty-pattern": require_no_empty_pattern.default,
		"no-export-in-script-setup": require_no_export_in_script_setup.default,
		"no-expose-after-await": require_no_expose_after_await.default,
		"no-extra-parens": require_no_extra_parens.default,
		"no-implicit-coercion": require_no_implicit_coercion.default,
		"no-import-compiler-macros": require_no_import_compiler_macros.default,
		"no-irregular-whitespace": require_no_irregular_whitespace.default,
		"no-lifecycle-after-await": require_no_lifecycle_after_await.default,
		"no-literals-in-template": require_no_literals_in_template.default,
		"no-lone-template": require_no_lone_template.default,
		"no-loss-of-precision": require_no_loss_of_precision.default,
		"no-multi-spaces": require_no_multi_spaces.default,
		"no-multiple-objects-in-class": require_no_multiple_objects_in_class.default,
		"no-multiple-slot-args": require_no_multiple_slot_args.default,
		"no-multiple-template-root": require_no_multiple_template_root.default,
		"no-mutating-props": require_no_mutating_props.default,
		"no-negated-condition": require_no_negated_condition.default,
		"no-negated-v-if-condition": require_no_negated_v_if_condition.default,
		"no-parsing-error": require_no_parsing_error.default,
		"no-potential-component-option-typo": require_no_potential_component_option_typo.default,
		"no-props-shadow": require_no_props_shadow.default,
		"no-ref-as-operand": require_no_ref_as_operand.default,
		"no-ref-object-reactivity-loss": require_no_ref_object_reactivity_loss.default,
		"no-required-prop-with-default": require_no_required_prop_with_default.default,
		"no-reserved-component-names": require_no_reserved_component_names.default,
		"no-reserved-keys": require_no_reserved_keys.default,
		"no-reserved-props": require_no_reserved_props.default,
		"no-restricted-block": require_no_restricted_block.default,
		"no-restricted-call-after-await": require_no_restricted_call_after_await.default,
		"no-restricted-class": require_no_restricted_class.default,
		"no-restricted-component-names": require_no_restricted_component_names.default,
		"no-restricted-component-options": require_no_restricted_component_options.default,
		"no-restricted-custom-event": require_no_restricted_custom_event.default,
		"no-restricted-html-elements": require_no_restricted_html_elements.default,
		"no-restricted-props": require_no_restricted_props.default,
		"no-restricted-static-attribute": require_no_restricted_static_attribute.default,
		"no-restricted-syntax": require_no_restricted_syntax.default,
		"no-restricted-v-bind": require_no_restricted_v_bind.default,
		"no-restricted-v-on": require_no_restricted_v_on.default,
		"no-root-v-if": require_no_root_v_if.default,
		"no-setup-props-reactivity-loss": require_no_setup_props_reactivity_loss.default,
		"no-shadow-native-events": require_no_shadow_native_events.default,
		"no-shared-component-data": require_no_shared_component_data.default,
		"no-side-effects-in-computed-properties": require_no_side_effects_in_computed_properties.default,
		"no-spaces-around-equal-signs-in-attribute": require_no_spaces_around_equal_signs_in_attribute.default,
		"no-sparse-arrays": require_no_sparse_arrays.default,
		"no-static-inline-styles": require_no_static_inline_styles.default,
		"no-template-key": require_no_template_key.default,
		"no-template-shadow": require_no_template_shadow.default,
		"no-template-target-blank": require_no_template_target_blank.default,
		"no-textarea-mustache": require_no_textarea_mustache.default,
		"no-this-in-before-route-enter": require_no_this_in_before_route_enter.default,
		"no-undef-components": require_no_undef_components.default,
		"no-undef-directives": require_no_undef_directives.default,
		"no-undef-properties": require_no_undef_properties.default,
		"no-unsupported-features": require_no_unsupported_features.default,
		"no-unused-components": require_no_unused_components.default,
		"no-unused-emit-declarations": require_no_unused_emit_declarations.default,
		"no-unused-properties": require_no_unused_properties.default,
		"no-unused-refs": require_no_unused_refs.default,
		"no-unused-vars": require_no_unused_vars.default,
		"no-use-computed-property-like-method": require_no_use_computed_property_like_method.default,
		"no-use-v-else-with-v-for": require_no_use_v_else_with_v_for.default,
		"no-use-v-if-with-v-for": require_no_use_v_if_with_v_for.default,
		"no-useless-concat": require_no_useless_concat.default,
		"no-useless-mustaches": require_no_useless_mustaches.default,
		"no-useless-template-attributes": require_no_useless_template_attributes.default,
		"no-useless-v-bind": require_no_useless_v_bind.default,
		"no-v-for-template-key-on-child": require_no_v_for_template_key_on_child.default,
		"no-v-for-template-key": require_no_v_for_template_key.default,
		"no-v-html": require_no_v_html.default,
		"no-v-model-argument": require_no_v_model_argument.default,
		"no-v-text-v-html-on-component": require_no_v_text_v_html_on_component.default,
		"no-v-text": require_no_v_text.default,
		"no-watch-after-await": require_no_watch_after_await.default,
		"object-curly-newline": require_object_curly_newline.default,
		"object-curly-spacing": require_object_curly_spacing.default,
		"object-property-newline": require_object_property_newline.default,
		"object-shorthand": require_object_shorthand.default,
		"one-component-per-file": require_one_component_per_file.default,
		"operator-linebreak": require_operator_linebreak.default,
		"order-in-components": require_order_in_components.default,
		"padding-line-between-blocks": require_padding_line_between_blocks.default,
		"padding-line-between-tags": require_padding_line_between_tags.default,
		"padding-lines-in-component-definition": require_padding_lines_in_component_definition.default,
		"prefer-define-options": require_prefer_define_options.default,
		"prefer-import-from-vue": require_prefer_import_from_vue.default,
		"prefer-prop-type-boolean-first": require_prefer_prop_type_boolean_first.default,
		"prefer-separate-static-class": require_prefer_separate_static_class.default,
		"prefer-single-event-payload": require_prefer_single_event_payload.default,
		"prefer-template": require_prefer_template.default,
		"prefer-true-attribute-shorthand": require_prefer_true_attribute_shorthand.default,
		"prefer-use-template-ref": require_prefer_use_template_ref.default,
		"prefer-v-model": require_prefer_v_model.default,
		"prop-name-casing": require_prop_name_casing.default,
		"quote-props": require_quote_props.default,
		"require-component-is": require_require_component_is.default,
		"require-default-export": require_require_default_export.default,
		"require-default-prop": require_require_default_prop.default,
		"require-direct-export": require_require_direct_export.default,
		"require-emit-validator": require_require_emit_validator.default,
		"require-explicit-emits": require_require_explicit_emits.default,
		"require-explicit-slots": require_require_explicit_slots.default,
		"require-expose": require_require_expose.default,
		"require-macro-variable-name": require_require_macro_variable_name.default,
		"require-name-property": require_require_name_property.default,
		"require-prop-comment": require_require_prop_comment.default,
		"require-prop-type-constructor": require_require_prop_type_constructor.default,
		"require-prop-types": require_require_prop_types.default,
		"require-render-return": require_require_render_return.default,
		"require-slots-as-functions": require_require_slots_as_functions.default,
		"require-toggle-inside-transition": require_require_toggle_inside_transition.default,
		"require-typed-object-prop": require_require_typed_object_prop.default,
		"require-typed-ref": require_require_typed_ref.default,
		"require-v-for-key": require_require_v_for_key.default,
		"require-valid-default-prop": require_require_valid_default_prop.default,
		"restricted-component-names": require_restricted_component_names.default,
		"return-in-computed-property": require_return_in_computed_property.default,
		"return-in-emits-validator": require_return_in_emits_validator.default,
		"script-indent": require_script_indent.default,
		"singleline-html-element-content-newline": require_singleline_html_element_content_newline.default,
		"slot-name-casing": require_slot_name_casing.default,
		"sort-keys": require_sort_keys.default,
		"space-in-parens": require_space_in_parens.default,
		"space-infix-ops": require_space_infix_ops.default,
		"space-unary-ops": require_space_unary_ops.default,
		"static-class-names-order": require_static_class_names_order.default,
		"template-curly-spacing": require_template_curly_spacing.default,
		"this-in-template": require_this_in_template.default,
		"use-v-on-exact": require_use_v_on_exact.default,
		"v-bind-style": require_v_bind_style.default,
		"v-for-delimiter-style": require_v_for_delimiter_style.default,
		"v-if-else-key": require_v_if_else_key.default,
		"v-on-event-hyphenation": require_v_on_event_hyphenation.default,
		"v-on-handler-style": require_v_on_handler_style.default,
		"v-on-style": require_v_on_style.default,
		"v-slot-style": require_v_slot_style.default,
		"valid-attribute-name": require_valid_attribute_name.default,
		"valid-define-emits": require_valid_define_emits.default,
		"valid-define-options": require_valid_define_options.default,
		"valid-define-props": require_valid_define_props.default,
		"valid-model-definition": require_valid_model_definition.default,
		"valid-next-tick": require_valid_next_tick.default,
		"valid-template-root": require_valid_template_root.default,
		"valid-v-bind-sync": require_valid_v_bind_sync.default,
		"valid-v-bind": require_valid_v_bind.default,
		"valid-v-cloak": require_valid_v_cloak.default,
		"valid-v-else-if": require_valid_v_else_if.default,
		"valid-v-else": require_valid_v_else.default,
		"valid-v-for": require_valid_v_for.default,
		"valid-v-html": require_valid_v_html.default,
		"valid-v-if": require_valid_v_if.default,
		"valid-v-is": require_valid_v_is.default,
		"valid-v-memo": require_valid_v_memo.default,
		"valid-v-model": require_valid_v_model.default,
		"valid-v-on": require_valid_v_on.default,
		"valid-v-once": require_valid_v_once.default,
		"valid-v-pre": require_valid_v_pre.default,
		"valid-v-show": require_valid_v_show.default,
		"valid-v-slot": require_valid_v_slot.default,
		"valid-v-text": require_valid_v_text.default
	},
	processors: {
		".vue": require_processor.default,
		vue: require_processor.default
	}
};
exports.default = plugin_default;
