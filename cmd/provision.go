package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"govpn/internal"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	startProvisionCommand = &cobra.Command{
		Use:   "start",
		Short: "Exec `start-session`",
		Long:  "Exec `start-session`",
		Run: func(_ *cobra.Command, _ []string) {
			var (
				ami                *internal.Ami
				az                 *internal.AvailabilityZone
				ec2Type            *internal.InstanceType
				defaultVpc         *internal.DefaultVpc
				err                error
				_terraformVarsJson = &TerraformVarsJson{}
			)
			ctx := context.Background()

			defaultVpc, err = internal.DefaultVpcExists(ctx, *_credential.awsConfig)
			if err != nil {
				panicRed(err)
			}

			if !defaultVpc.Existence {
				answer, err := internal.AskCreateDefaultVpc()
				if err != nil {
					panicRed(err)
				}

				if answer == "Yes" {
					vpc, err := internal.CreateDefaultVpc(ctx, *_credential.awsConfig)
					if err != nil {
						panicRed(err)
					}
					internal.PrintReady("create-vpc", _credential.awsConfig.Region, "vpc-id", vpc.Id)
				} else {
					os.Exit(1)
				}
			}

			if _, err := os.Stat(_defaultTerraformVars); err == nil {

				buffer, err := os.ReadFile(_defaultTerraformVars)
				if err != nil {
					panicRed(err)
				}
				json.NewDecoder(bytes.NewBuffer(buffer)).Decode(&_terraformVarsJson)

				answer, err := internal.AskNewTfVars(_terraformVarsJson.Aws_Region, _terraformVarsJson.Availability_Zone, _terraformVarsJson.Instance_Type, _terraformVarsJson.Ec2_Ami)
				if err != nil {
					panicRed(err)
				}

				if strings.Split(answer, ",")[0] == "No" {

					_terraformVarsJson.Aws_Region = _credential.awsConfig.Region
					notice := color.New(color.Bold, color.FgHiRed).PrintfFunc()
					notice("⚠️\tIf you want to change region, restart `govpn start -r {region}`\n")

					if az == nil {
						az, err = internal.AskAvailabilityZone(ctx, *_credential.awsConfig)
						if err != nil {
							panicRed(err)
						}
						_terraformVarsJson.Availability_Zone = az.Name
					}
					internal.PrintReady("[start-provisioning]", _credential.awsConfig.Region, "availability-zone", az.Name)

					if ec2Type == nil {
						ec2Type, err = internal.AskInstanceType(ctx, *_credential.awsConfig, az.Name)
						if err != nil {
							panicRed(err)
						}
						_terraformVarsJson.Instance_Type = ec2Type.Name
					}
					internal.PrintReady("[start-provisioning]", _credential.awsConfig.Region, "instance-type", ec2Type.Name)

					if ami == nil {
						ami, err = internal.AskAmi(ctx, *_credential.awsConfig)
						if err != nil {
							panicRed(err)
						}
						_terraformVarsJson.Ec2_Ami = ami.Name
					}
					internal.PrintReady("[start-provisioning]", _credential.awsConfig.Region, "ami-id", ami.Name)

					// save tfvars ===============================
					jsonData := make(map[string]interface{})
					jsonData["aws_region"] = _terraformVarsJson.Aws_Region
					jsonData["ec2_ami"] = _terraformVarsJson.Ec2_Ami
					jsonData["instance_type"] = _terraformVarsJson.Instance_Type
					jsonData["availability_zone"] = _terraformVarsJson.Availability_Zone

					if _, err = internal.SaveTerraformVariable(jsonData, _defaultTerraformVars); err != nil {
						panicRed(err)
					}
				}
			} else {
				_terraformVarsJson.Aws_Region = _credential.awsConfig.Region
				notice := color.New(color.Bold, color.FgHiRed).PrintfFunc()
				notice("⚠️\t\tIf you want to change region, restart `govpn start -r {region}`\n")

				if az == nil {
					az, err = internal.AskAvailabilityZone(ctx, *_credential.awsConfig)
					if err != nil {
						panicRed(err)
					}
					_terraformVarsJson.Availability_Zone = az.Name
				}
				internal.PrintReady("[start-provisioning]", _credential.awsConfig.Region, "availability-zone", az.Name)

				if ec2Type == nil {
					ec2Type, err = internal.AskInstanceType(ctx, *_credential.awsConfig, az.Name)
					if err != nil {
						panicRed(err)
					}
					_terraformVarsJson.Instance_Type = ec2Type.Name
				}
				internal.PrintReady("[start-provisioning]", _credential.awsConfig.Region, "instance-type", ec2Type.Name)

				if ami == nil {
					ami, err = internal.AskAmi(ctx, *_credential.awsConfig)
					if err != nil {
						panicRed(err)
					}
					_terraformVarsJson.Ec2_Ami = ami.Name
				}
				internal.PrintReady("[start-provisioning]", _credential.awsConfig.Region, "ami-id", ami.Name)

				// save tfvars ===============================
				jsonData := make(map[string]interface{})
				jsonData["aws_region"] = _terraformVarsJson.Aws_Region
				jsonData["ec2_ami"] = _terraformVarsJson.Ec2_Ami
				jsonData["instance_type"] = _terraformVarsJson.Instance_Type
				jsonData["availability_zone"] = _terraformVarsJson.Availability_Zone

				if _, err = internal.SaveTerraformVariable(jsonData, _defaultTerraformVars); err != nil {
					panicRed(err)
				}
			}

			// terraform ready ===============================
			execPath, err := internal.TerraformReady(ctx, "1.0.6")
			if err != nil {
				panicRed(err)
			}
			tf, err := internal.SetRoot(execPath, _defaultTerraformPath)
			if err != nil {
				panicRed(err)
			}
			internal.PrintReady("[start-provisioning]", _credential.awsConfig.Region, "terraform-state", "ready")

			// terraform init ===============================
			if err = tf.Init(context.Background(), tfexec.Upgrade(true)); err != nil {
				panicRed(fmt.Errorf("[err] failed to terraform init"))
			}
			internal.PrintReady("[start-provisioning]", _credential.awsConfig.Region, "terraform init", "success")

			// terraform plan ===============================
			if _, err = tf.Plan(context.Background(),
				tfexec.VarFile(_defaultTerraformVars)); err != nil {
				panicRed(fmt.Errorf("[err] failed to terraform plan"))
			}
			internal.PrintReady("[start-provisioning]", _credential.awsConfig.Region, "terraform plan", "success")

			// terraform apply ===============================
			answer, err := internal.AskTerraformApply()
			if err != nil {
				panicRed(err)
			}

			if answer == "Yes" {
				if err = internal.TerraformApply(_defaultTerraformPath); err != nil {
					panicRed(fmt.Errorf("[err] failed to terraform apply"))
				}

				congratulation := color.New(color.Bold, color.FgHiGreen).PrintFunc()
				congratulation("🎉 Provisioning Complete! 🎉")
			}
		},
	}
)

func init() {
	startProvisionCommand.Flags().StringP("ami", "a", "", "")
	viper.BindPFlag("start-session-target", startProvisionCommand.Flags().Lookup("ami"))

	rootCmd.AddCommand(startProvisionCommand)
}
