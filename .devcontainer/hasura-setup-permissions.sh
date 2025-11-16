#!/bin/bash
# Setup Hasura permissions for production

HASURA_URL="${HASURA_URL:-http://localhost:8080}"
HASURA_SECRET="${HASURA_ADMIN_SECRET:-dev-admin-secret}"

echo "Setting up Hasura permissions..."

# Function to call Hasura metadata API
call_metadata_api() {
  curl -X POST "${HASURA_URL}/v1/metadata" \
    -H "Content-Type: application/json" \
    -H "x-hasura-admin-secret: ${HASURA_SECRET}" \
    -d "$1"
}

# 1. Create user role permissions for goals table
echo "Setting up goals table permissions..."

# Drop existing permissions first
call_metadata_api '{
  "type": "pg_drop_select_permission",
  "args": {
    "table": "goals",
    "role": "user"
  }
}'

call_metadata_api '{
  "type": "pg_drop_insert_permission",
  "args": {
    "table": "goals",
    "role": "user"
  }
}'

call_metadata_api '{
  "type": "pg_drop_update_permission",
  "args": {
    "table": "goals",
    "role": "user"
  }
}'

call_metadata_api '{
  "type": "pg_drop_delete_permission",
  "args": {
    "table": "goals",
    "role": "user"
  }
}'

# Create new permissions
call_metadata_api '{
  "type": "pg_create_select_permission",
  "args": {
    "table": "goals",
    "role": "user",
    "permission": {
      "columns": ["id", "title", "created", "finished"],
      "filter": {
        "user_goals": {
          "user_id": {
            "_eq": "X-Hasura-User-Id"
          }
        }
      }
    }
  }
}'

call_metadata_api '{
  "type": "pg_create_insert_permission",
  "args": {
    "table": "goals",
    "role": "user",
    "permission": {
      "check": {},
      "columns": ["title", "finished"]
    }
  }
}'

call_metadata_api '{
  "type": "pg_create_update_permission",
  "args": {
    "table": "goals",
    "role": "user",
    "permission": {
      "columns": ["title", "finished"],
      "filter": {
        "user_goals": {
          "user_id": {
            "_eq": "X-Hasura-User-Id"
          }
        }
      }
    }
  }
}'

call_metadata_api '{
  "type": "pg_create_delete_permission",
  "args": {
    "table": "goals",
    "role": "user",
    "permission": {
      "filter": {
        "user_goals": {
          "user_id": {
            "_eq": "X-Hasura-User-Id"
          }
        }
      }
    }
  }
}'

# 2. Create user role permissions for user_goals table
echo "Setting up user_goals table permissions..."

# Drop existing permissions
call_metadata_api '{
  "type": "pg_drop_select_permission",
  "args": {
    "table": "user_goals",
    "role": "user"
  }
}'

call_metadata_api '{
  "type": "pg_drop_insert_permission",
  "args": {
    "table": "user_goals",
    "role": "user"
  }
}'

# Create new permissions
call_metadata_api '{
  "type": "pg_create_select_permission",
  "args": {
    "table": "user_goals",
    "role": "user",
    "permission": {
      "columns": ["user_id", "goal_id"],
      "filter": {
        "user_id": {
          "_eq": "X-Hasura-User-Id"
        }
      }
    }
  }
}'

call_metadata_api '{
  "type": "pg_create_insert_permission",
  "args": {
    "table": "user_goals",
    "role": "user",
    "permission": {
      "check": {
        "user_id": {
          "_eq": "X-Hasura-User-Id"
        }
      },
      "columns": ["goal_id"]
    }
  }
}'

# 3. Create user role permissions for goal_relations table
echo "Setting up goal_relations table permissions..."

# Drop existing permissions
call_metadata_api '{
  "type": "pg_drop_select_permission",
  "args": {
    "table": "goal_relations",
    "role": "user"
  }
}'

call_metadata_api '{
  "type": "pg_drop_insert_permission",
  "args": {
    "table": "goal_relations",
    "role": "user"
  }
}'

call_metadata_api '{
  "type": "pg_drop_delete_permission",
  "args": {
    "table": "goal_relations",
    "role": "user"
  }
}'

# Create new permissions
call_metadata_api '{
  "type": "pg_create_select_permission",
  "args": {
    "table": "goal_relations",
    "role": "user",
    "permission": {
      "columns": ["parent_id", "child_id"],
      "filter": {
        "_and": [
          {
            "goalByParentId": {
              "user_goals": {
                "user_id": {
                  "_eq": "X-Hasura-User-Id"
                }
              }
            }
          },
          {
            "goal": {
              "user_goals": {
                "user_id": {
                  "_eq": "X-Hasura-User-Id"
                }
              }
            }
          }
        ]
      }
    }
  }
}'

call_metadata_api '{
  "type": "pg_create_insert_permission",
  "args": {
    "table": "goal_relations",
    "role": "user",
    "permission": {
      "check": {
        "_and": [
          {
            "goalByParentId": {
              "user_goals": {
                "user_id": {
                  "_eq": "X-Hasura-User-Id"
                }
              }
            }
          },
          {
            "goal": {
              "user_goals": {
                "user_id": {
                  "_eq": "X-Hasura-User-Id"
                }
              }
            }
          }
        ]
      },
      "columns": ["parent_id", "child_id"]
    }
  }
}'

call_metadata_api '{
  "type": "pg_create_delete_permission",
  "args": {
    "table": "goal_relations",
    "role": "user",
    "permission": {
      "filter": {
        "_and": [
          {
            "goalByParentId": {
              "user_goals": {
                "user_id": {
                  "_eq": "X-Hasura-User-Id"
                }
              }
            }
          },
          {
            "goal": {
              "user_goals": {
                "user_id": {
                  "_eq": "X-Hasura-User-Id"
                }
              }
            }
          }
        ]
      }
    }
  }
}'

# 4. Setup user table permissions (read only own user)
echo "Setting up users table permissions..."

# Drop existing permission
call_metadata_api '{
  "type": "pg_drop_select_permission",
  "args": {
    "table": "users",
    "role": "user"
  }
}'

# Create new permission
call_metadata_api '{
  "type": "pg_create_select_permission",
  "args": {
    "table": "users",
    "role": "user",
    "permission": {
      "columns": ["id", "sub", "email", "first_name", "last_name", "created"],
      "filter": {
        "id": {
          "_eq": "X-Hasura-User-Id"
        }
      }
    }
  }
}'

echo "✅ Hasura permissions setup complete!"
