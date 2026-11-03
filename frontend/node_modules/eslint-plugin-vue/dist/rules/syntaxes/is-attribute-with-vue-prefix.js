// lib/rules/syntaxes/is-attribute-with-vue-prefix.ts
/**
* @author Yosuke Ota
* See LICENSE file in root directory for full license.
*/
var is_attribute_with_vue_prefix_default = {
	supported: ">=3.1.0",
	createTemplateBodyVisitor(context) {
		return { "VAttribute[directive=false][key.name='is']"(node) {
			if (!node.value) return;
			if (node.value.value.startsWith("vue:")) context.report({
				node: node.value,
				messageId: "forbiddenIsAttributeWithVuePrefix"
			});
		} };
	}
};
exports.default = is_attribute_with_vue_prefix_default;
