init the frontend app, using latest version of Next.js
create a frontend folder and put the frontend code inside it
for the styling, use latest tailwindcss
render the frontend app using CSR mode
don't use shadcn-ui or any other UI library, instead use custom components and styles
always create components like button, input, card, etc.
for now let's use the tailwind color palette
but in the future, I'll use a custom color palette
the frontend app mostly for admin panel
it will cover login, register, reset password, and dashboard
add sidebar menu with tree structure
example sidebar structure:
- Master Data
  - Product
  - Category
  - Supplier
- Transaction
  - Sales
  - Purchase
- Report
  - Sales Report
  - Purchase Report
add a header with logo and user name, if user name is clicked, show dropdown menu with options to edit profile, change password, and logout
add the simple footer with copyright information

for phase 1, build pages:
1. login
2. register
3. reset password
4. dashboard
5. master category

all buttons should be clickable
add the alert, success, or error message
for example in the login page, if the user clicks the login button, show a success message
if the user clicks the register button, show error messsage if the email is already registered
show the example data in the master category page, use json file or maybe simple variable for the example data
