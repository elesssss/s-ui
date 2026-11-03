// lib/rules/syntaxes/v-bind-prop-modifier-shorthand.ts
/**
* @author Yosuke Ota
* See LICENSE file in root directory for full license.
*/
var v_bind_prop_modifier_shorthand_default = {
	supported: ">=3.2.0 || >=2.6.0-beta.1 <=2.6.0-beta.3",
	createTemplateBodyVisitor(context) {
		/**
		* Reports `.prop` shorthand node
		*/
		function reportPropModifierShorthand(bindPropKey) {
			context.report({
				node: bindPropKey,
				messageId: "forbiddenVBindPropModifierShorthand",
				fix: (fixer) => fixer.replaceText(bindPropKey, `:${bindPropKey.argument.rawName}.prop`)
			});
		}
		return { "VAttribute[directive=true] > VDirectiveKey[name.name='bind'][name.rawName='.']": reportPropModifierShorthand };
	}
};
exports.default = v_bind_prop_modifier_shorthand_default;
