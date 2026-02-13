1. in the table, I want to add the option to sort by selected column, now because we don't have backend yet, we can use the example data to sort the table, but in the future the sorting must be done by the backend.
2. in the pagination I want to add the option to change the number of items per page, now because we don't have backend yet, we can use the example data to change the number of items per page, but in the future the pagination must be done by the backend.
3. add the master data product, this is the complex thing to implement. here the requirements:
  - product will have a name, description, multi images, a category.
  - product will have multiple units & conversion. because every products will have different units and packaging.
    i.e:
    - 1 Ton = 100 Kuintal
    - 1 Kuintal = 100 Kilogram
    - 1 Kilogram = 10 Ons
    - 1 Ons = 100 Gram
    
    - 1 Carton = 10 Boxes
    - 1 Boxes = 12 Pieces
    
    the unit is flexible, for example when the user purchase from supplier it can use higher unit, for example 10 Boxes, but when selling it can use lower unit, for example 12 Pieces. or for slow moving products, it can use lower unit, user can also purchase in smaller units like pieces. also for wholesale products, user can sell it use larger units like cartons. so it's pretty flexible. I'm still not sure about the best way to implement this, give me some ideas and recommendation solutions.
  - product will have price settings, fixed price or markup price (markup can be percentage or fixed amount)
    this setting will be affected in the variants, if use choose fixed price, in the variant will have input to fill the fixed price, if use choose markup price, in the variant will have input to fill the markup percentage or fixed amount.
  - product will have multiple variants, variant can be combination.
    i.e:
    - Variant 1
      - color: blue
      - size: small
    - Variant 2
      - color: red
      - size: medium
    - Variant 3
      - color: green
      - size: large
    each variants will have different price
    price in the variant will have options to choose wholesale price or retail price
    if choose wholesale price, in the variant will have input to fill the wholesale rule
     - condition | QTY | price by product configuration, if choose fixed price then input the fixed price, if choose markup price then input the markup percentage or fixed amount.
     i.e for fixed price:
     - > | 1 | 1000
     - > | 10 | 900
     - > | 100 | 800
     i.e for markup price percentage:
     - > | 1 | 10%
     - > | 10 | 5%
     - > | 100 | 2%
     i.e for markup price fixed nominal/amount:
     - > | 1 | 1000
     - > | 10 | 500
     - > | 100 | 200
    if choose retail price, in the variant will have input to fill the retail price
  variant is optional, in the product have option or radio checkbox to select "has variant" or not.
  if not has variant show the simple form input
  if has variant show the variants form input
  but I want to implement the both option in the one database schema (later for the backend & frontend integration)
  but now the most important in the frontend is show the simple form vs variants form input
  maybe add tabs in the bottom of the product fields.
  Units | Variants
