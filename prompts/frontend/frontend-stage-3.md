Add one more menu Settings
the menu will have two sub menus:
 - Users
 - Roles & Permissions

1. users 
  - users will save the name, email, phone, address, password, profile picture
  - super admin can create a user, after that super admin can share the email and default password to the user
  - when create a new user, show the select options for roles (multiple, nullable), save it to zustand state or whatever, show success message and error message if any, don't show the password, just show notification if credentials are sent to user email
  - or user can register themselves, but super admin will verify it before they can login
  - in the users can choose multiple roles
  - show the users in the table, add sorting, pagination, and search, follow the current table design, layout and style
  - when click create & edit user show the popup form
  - when click delete user show the confirmation popup form
  - add the predefined mock data for users, for now mock the data for super admin user, it can't be deleted
  - mock the five more users data, except super admin role
2. roles & permissions
  - example role is: admin, cashier, manager, accountant, warehouse etc.
  - role can be assigned to user
  - user can have multiple roles
  - show the success message after creating a role
  - in the role table, have Permissions button, when clicked it will show all the permissions
  - role can have multiple permissions
  - example permission is: Read, Create, Update, Delete, Export. but it's depends on the feature, maybe some feature doesn't have Delete or Export permission
  - list permissions will displayed as a tree view, i.e:
    - Master Data
      - Product    checkbox -> | Read | Create | Update | Delete | Export
    - Transaction
      - Sale     checkbox -> | Read | Create | Update | Delete | Export
      - Purchase checkbox -> | Read | Create | Update | Delete | Export
  - the permissions doesn't have CRUD feature, it's the seeder data from the engineer, and it's depends on the feature, if developers add more feature, then they will add more permissions for that feature
  - add the predefined mock data for roles and permissions
  - but for now don't implement the permissions feature limitation, keep the feature and menu accessible for all users (will be implemented in the next stage with the backend API)
