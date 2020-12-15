provider "aws" {
  region = "us-west-2"
  profile = "default"
}

module "eks" {
  source = "../../../../modules/eks"

  vpc_name = "kevin"
  cluster_name = "kevin"
  subnet_cidrs = ["10.0.3.0/24", "10.0.4.0/24"]
  subnet_azs = ["us-west-2c", "us-west-2b"]
}
