// lib/rules/syntaxes/scope-attribute.ts
/**
* @author Yosuke Ota
* See LICENSE file in root directory for full license.
*/
var scope_attribute_default = {
	deprecated: "2.5.0",
	supported: "<3.0.0",
	createTemplateBodyVisitor(context) {
		/**
		* Reports `scope` node
		*/
		function reportScope(scopeKey) {
			context.report({
				node: scopeKey,
				messageId: "forbiddenScopeAttribute",
				fix: (fixer) => fixer.replaceText(scopeKey, "slot-scope")
			});
		}
		return { "VAttribute[directive=true] > VDirectiveKey[name.name='scope']": reportScope };
	}
};
exports.default = scope_attribute_default;
