I want to work on the /transaction/sales page.
This is not connected with the backend, mock the data and store it in the zustand.
In this page, user will sell the products and variants to buyers.
I want to add the multi session sales page, where users can sell products and variants to multiple buyers simultaneously and can switch between sessions without any interruption.
User can close the session by clicking on the close button, make the seamless confirmation when closing the session, one click to change the icon to confirm, and second click to approve the confirmation.
User can search the products and variants by entering the name, SKU or barcode, add button to search.
User must type at least 3 characters to start the search, handle enter to search.
The search results will be displayed in a dropdown with columns for product & all variants image, name, description, SKU, unit.
If the search results are empty, display a message saying "No results found".
Limit the search results to 10 items.
The dropdown will look like this:
  Main Image | Product Name
    - Main Image | SKU 1 | Variant attributes | stock | Button to select
    - Main Image | SKU 2 | Variant attributes | stock | Button to select
  Main Image | Product Name
    - Main Image | SKU 1 | Variant attributes | stock | Button to select
    - Main Image | SKU 2 | Variant attributes | stock | Button to select
    - Main Image | SKU 3 | Variant attributes | stock | Button to select
If the stock is 0, make the variant background color soft red. and can't be selected.
If the user selects a variant, the variant will be added to the cart. don't close the dropdown until user closes it. add button to close the dropdown.
Set the default quantity to 1. don't allow the user to change quantity to less than 1.
If the price have tiered pricing, and the quantity is meet with the tiered quantity condition, automatically update and show the tiered price.
If the stock is less than the quantity, show an error message.
In the cart, show the variant image, name, desc, sku, price, stock
user can:
  - edit the quantity
  - change the unit
  - remove the variants from the cart
  - and proceed to checkout
The example cart will look like this:
  Variant Image | SKU | Name | Attributes | Quantity | Unit (default is base_unit) | Price | Total | Actions <br>
                | Stock | Description |
  Variant Image | SKU | Name | Attributes | Quantity | Unit (default is base_unit) | Price | Total | Actions <br>
                | Stock | Description |
After the product lists, in the below show the cart summary:
  - Total Items: 2
  - Subtotal: $10.00
  - Grand Total: $10.00
Show the payment method selection form. Cash, Card or QRIS
Then add button to proceed to checkout. show the success message after checkout.
After choose payment method and checkout, show the receipt, give option to print or save as PDF.
