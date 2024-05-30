resource "render_project" "my-project" {
  name = "my project"
  environments = {
    "development" : {
      name : "development",
      protected_status : "unprotected"
    },
    "staging" : {
      name : "staging",
      protected_status : "protected"
    },
    "production" : {
      name : "production",
      protected_status : "unprotected"
    },
  }
}