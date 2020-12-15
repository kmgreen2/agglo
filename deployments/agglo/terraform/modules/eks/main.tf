data "aws_vpc" "main" {
  tags = {
    Name = var.vpc_name
  }
}

data "aws_nat_gateway" "nat" {
  tags = {
    Name = "${var.vpc_name}-nat"
  }
}

resource "aws_eks_cluster" "default" {
  name     = var.cluster_name
  role_arn = aws_iam_role.default.arn

  vpc_config {
    subnet_ids = aws_subnet.cluster[*].id
  }

  # Ensure that IAM Role permissions are created before and deleted after EKS Cluster handling.
  # Otherwise, EKS will not be able to properly delete EKS managed EC2 infrastructure such as Security Groups.
  depends_on = [
    aws_iam_role_policy_attachment.default-AmazonEKSClusterPolicy,
    aws_iam_role_policy_attachment.default-AmazonEKSVPCResourceController,
  ]
}

resource "aws_eks_node_group" "default" {
  cluster_name    = aws_eks_cluster.default.name
  node_group_name = "${var.cluster_name}-nodes"
  node_role_arn   = aws_iam_role.default.arn
  subnet_ids      = aws_subnet.cluster[*].id

  instance_types = ["t2.2xlarge"]

  remote_access {
    ec2_ssh_key = var.ec2_ssh_key
  }

  scaling_config {
    desired_size = 2 
    max_size     = 2
    min_size     = 1
  }

  # Ensure that IAM Role permissions are created before and deleted after EKS Node Group handling.
  # Otherwise, EKS will not be able to properly delete EC2 Instances and Elastic Network Interfaces.
  depends_on = [
    aws_iam_role_policy_attachment.default-AmazonEKSWorkerNodePolicy,
    aws_iam_role_policy_attachment.default-AmazonEKS_CNI_Policy,
    aws_iam_role_policy_attachment.default-AmazonEC2ContainerRegistryReadOnly,
  ]
}

resource "aws_subnet" "cluster" {
  count = length(var.subnet_cidrs)
  vpc_id     = data.aws_vpc.main.id
  cidr_block = var.subnet_cidrs[count.index]
  availability_zone = var.subnet_azs[count.index]

  tags = {
    Name = "${var.cluster_name} subnet ${count.index}"
    "kubernetes.io/cluster/${var.cluster_name}" = "shared"
  }
}

resource "aws_route_table" "nat" {
  vpc_id = data.aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0"
    nat_gateway_id = data.aws_nat_gateway.nat.id
  }
}

resource "aws_route_table_association" "instance" {
  count = length(var.subnet_cidrs)
  subnet_id = aws_subnet.cluster[count.index].id
  route_table_id = aws_route_table.nat.id
}

resource "aws_iam_role" "default" {
  name = "eks-${var.cluster_name}-default"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    },
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "default-AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.default.name
}

# Optionally, enable Security Groups for Pods
# Reference: https://docs.aws.amazon.com/eks/latest/userguide/security-groups-for-pods.html
resource "aws_iam_role_policy_attachment" "default-AmazonEKSVPCResourceController" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController"
  role       = aws_iam_role.default.name
}

resource "aws_iam_role_policy_attachment" "default-AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.default.name
}

resource "aws_iam_role_policy_attachment" "default-AmazonEKS_CNI_Policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.default.name
}

resource "aws_iam_role_policy_attachment" "default-AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.default.name
}
