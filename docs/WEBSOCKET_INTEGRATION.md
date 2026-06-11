# WebSocket Integration Guide

## 1. WebSocket Endpoints

Connect to one of the following endpoints (replace `localhost:3000` with your server address if needed):

- ws://localhost:3000/ws/inventory
- ws://localhost:3000/ws/warehouses
- ws://localhost:3000/ws/products
- ws://localhost:3000/ws/suppliers

## 2. Connection Example (JavaScript)

```js
const ws = new WebSocket('ws://localhost:3000/ws/inventory');

ws.onopen = () => {
  console.log('WebSocket connected');
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  // Handle data based on type
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket disconnected');
};
```

## 3. Message Types & Payloads

You will receive JSON messages with a `type` field. Main types:

### a. Inventory Update

```json
{
  "type": "inventory_update",
  "inventory_id": 1,
  "product_id": 101,
  "warehouse_id": 5,
  "quantity": 50,
  "action": "created|updated|deleted|adjusted",
  "timestamp": "2026-02-14T12:34:56Z"
}
```

### b. Low Stock Alert

```json
{
  "type": "low_stock_alert",
  "product_id": 101,
  "warehouse_id": 5,
  "current_quantity": 3,
  "min_stock": 10,
  "product_name": "Widget",
  "timestamp": "2026-02-14T12:34:56Z"
}
```

### c. Warehouse/Product/Supplier Updates & Alerts

Similar structure, with fields like `type`, `*_id`, `name`, `action`, and `timestamp`.

## 4. Authentication

No authentication is required for WebSocket connections by default. (If you need to secure it, add token logic.)

## 5. Connection Lifecycle

- Connect using the endpoint.
- Listen for messages and handle by `type`.
- Reconnect logic is recommended if the connection drops.

## 6. Example Usage

See `web/inventory-realtime.html` for a full working example.
