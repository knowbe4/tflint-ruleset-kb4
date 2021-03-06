package rules

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TerraformDocumentedVariablesRule checks whether variables have descriptions
type TerraformValidatedVariablesRule struct {
	tflint.DefaultRule
}

// NewTerraformValidatedVariablesRule returns a new rule
func NewTerraformValidatedVariablesRule() *TerraformValidatedVariablesRule {
	return &TerraformValidatedVariablesRule{}
}

// Name returns the rule name
func (r *TerraformValidatedVariablesRule) Name() string {
	return "terraform_validated_variables"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformValidatedVariablesRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformValidatedVariablesRule) Severity() tflint.Severity {
	return tflint.ERROR
}

// Link returns the rule reference link
func (r *TerraformValidatedVariablesRule) Link() string {
	return "https://engineering.internal.knowbe4.com/tech-stack/terraform/style-guide/#validation"
}

// Check checks whether variables have descriptions
func (r *TerraformValidatedVariablesRule) Check(runner tflint.Runner) error {

	files, _ := runner.GetFiles()

	for filename := range files {
		r.checkFileSchema(runner, files[filename])
	}

	return nil
}

func (r *TerraformValidatedVariablesRule) isIgnoredType(block *hcl.Block) bool {

	// We ignore krn
	if block.Labels[0] == "krn" {
		return true
	}

	body, _, _ := block.Body.PartialContent(&hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     "type",
				Required: true,
			},
		},
	})

	// Traversal
	_, has_type := body.Attributes["type"]

	if has_type {
		for _, trav := range body.Attributes["type"].Expr.Variables() {
			// Ignore bool types because it doesnt make sense to validate them
			if trav.RootName() == "bool" {
				return true
			}
		}

	}

	return false
}

func (r *TerraformValidatedVariablesRule) checkFileSchema(runner tflint.Runner, file *hcl.File) error {

	content, _, diags := file.Body.PartialContent(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "variable",
				LabelNames: []string{"name"},
			},
		},
	})

	if diags.HasErrors() {
		return diags
	}

	for _, block := range content.Blocks.OfType("variable") {
		if r.isIgnoredType(block) {
			continue
		}

		c, _, _ := block.Body.PartialContent(&hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type: "validation",
				},
			},
		})

		if len(c.Blocks) == 0 {
			runner.EmitIssue(
				r,
				fmt.Sprintf("`%v` variable has no validations. Please include at least 1 validation for types that are not a bool.", block.Labels[0]),
				block.DefRange,
			)
		}
	}

	return nil
}
