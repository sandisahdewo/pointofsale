Note: backend is not implemented yet, add and store the mock data in the zustand
1. add master data supplier
  - save the name, address, phone number (optional), email (optional), and website (optional), bank accounts (optional, only store account name & account number, multiple) active (default true)
  - add the form validation and show error message if any error
  - show the success or error message after saving the supplier
2. in the product form (not variant), add the supplier field, can select multiple active suppliers, this is optional, to give the user flexibility
4. add master data Rack
  - save the name, code, location, capacity, active, description (optional)
  - this is for storing or displaying the products in the store
  - add the form validation and show error message if any error
  - show the success or error message after saving the rack
5. each variant add the Rack option, can select multiple active racks, this is optional
6. in the variant, remove the Price Type field, now only support Tiered pricing, if user want to sell as retail they can set only one price (Min Qty = 1), add note on the frontend so user will know this behavior.
7. in the product add Tab in the begining, tab name is Price
  - move the Price Setting & Markup Type there
  - add the price field based on the price setting and markup type, this is optional
  - add the Wholesale price field only, because I will remove the Price Type field in the variant (I described in the different part), Min Qty field and Sell Price or Markup field depends on the selected price setting
  - if price field is filled, it will applied to all the variants, price filled can be override in the variant, but even the variants have override value and user change the price field, it will replace the variant override value. add this note to the user so user will know this behavior. add the backend note for the future implementation
6. purchase order
  - in this step user can select one supplier (for purchase order destination)
  - add the date picker when the purchase order is created, default is today's date
  - add the status field (draft, sent, received, completed), default is draft
  - user can change the status to sent, it will send the purchase order to the supplier via email or whatsapp (later)
  - user can change the status to received, and start checking the received products
  - user can change the status to completed, it will mark the purchase order as completed
  - in the details, show the list of products (should be match with the selected supplier in product, product A select supplier 1) or show the list of products who don't have suppliers and the quantity is less than or equal to forecast minimum quantity, but user also can add more products even if the quantity is greater than the forecast minimum quantity
    - Product A
      - Variant 1 | Current Stock (read only) | Quantity: 0 (how much order qty, default is forecast minimum quantity) | Price: (read only, latest price from this supplier, if not available set as 0)
      - Variant 2 | Current Stock (read only) | Quantity: 0 (how much order qty, default is forecast minimum quantity) | Price: (read only, latest price from this supplier, if not available set as 0)
    - Product B
      - Variant 1 | Current Stock (read only) | Quantity: 0 (how much order qty, default is forecast minimum quantity) | Price: (read only, latest price from this supplier, if not available set as 0)
      - Variant 2 | Current Stock (read only) | Quantity: 0 (how much order qty, default is forecast minimum quantity) | Price: (read only, latest price from this supplier, if not available set as 0)
  - in the action show the Receive button, when clicked then:
    - update the status to received
    - add the received date & time, default is current date & time
    - add the payment method (dropdown, options: cash, credit card, bank transfer), default is cash
    - if payment method is not cash, show the bank account number field
    - show the subtotal, total price, total quantity items (count these fields from received quantity, this will be used to compare with the supplier bill)
    - show the ordered products & variants, with the ordered quantity and price
    - if the ordered quantity and received quantity are equal, then user can mark the product & variant as ok by clicking the checkbox and add warning message "Received quantity matches ordered quantity", but this warning message should be displayed only once or have a checkbox to mark as understand & don't show again and disable the input field
    - but if the received quantity is less than or greater than ordered quantity user can input the received quantity and remove the checkmark button
    - if the price is not equal to the latest price, user can input the new price and remove the checkmark button
    - then if all ok, then mark the purchase order as received and update the stock maybe have one button to save it in the bottom

Update:
1. in the product > Price tab, when change the sell price it's automatically showing the modal confirmation, with this implementation it's hard to edit and type new price. I want to change it, add a button to edit and save the new price, then if save button clicked it will show the modal confirmation.
2. in the purchase order items, I want to add the unit option, get the unit from the product, add the unit option before order qty field. when order user can choose the unit from the dropdown list. then the price will be calculated based on the selected unit (later, add the note for future)
3. in the purchase order receive step, show the selected unit before ordered quantity, then it will show the Variant | Selected Unit | Ordered Quantity | Received Quantity | Price
4. also in the purchase order view, show the selected unit too after SKU field
