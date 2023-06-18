terraform {
  required_providers {
    protomesh = {
      source = "protomesh/protomesh"
    }
    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
    }

  }
}

provider "random" {}

provider "protomesh" {
  address    = "localhost:6680"
  enable_tls = true
}

resource "random_uuid" "ping_pong_do_pong" {
}

resource "protomesh_gateway_policy" "ping_pong_do_pong" {

  namespace = "gateway"

  resource_id = random_uuid.ping_pong_do_pong.result
  name        = "Route to PingPong.DoPong method"

  policy {

    source {
      grpc {
        method_name = "/protomesh.services.v1.PingPongService/DoPing"
        exact_method_name_match = true
      }
    }
    
    handlers {
      handler {
        aws {
          handler {
            lambda {
              function_name = "protomeshPingback"
              qualifier     = "$LATEST"
            }
          }
        }
      }
    }

  }

}
