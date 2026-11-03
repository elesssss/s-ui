// lib/rules/syntaxes/dynamic-directive-arguments.ts
/**
* @author Yosuke Ota
* See LICENSE file in root directory for full license.
*/
var dynamic_directive_arguments_default = {
	supported: ">=2.6.0",
	createTemplateBodyVisitor(context) {
		/**
		* Reports dynamic argument node
		*/
		function reportDynamicArgument(dynamicArgument) {
			context.report({
				node: dynamicArgument,
				messageId: "forbiddenDynamicDirectiveArguments"
			});
		}
		return { "VAttribute[directive=true] > VDirectiveKey > VExpressionContainer": reportDynamicArgument };
	}
};
exports.default = dynamic_directive_arguments_default;
