const fetch = require("node-fetch");

const BASE_URL = "http://localhost:8082/api/v1";

async function testJWTAuth() {
  console.log("Testing JWT Authentication...\n");

  // Test 1: Try to access protected route without token
  console.log("1. Testing categories endpoint without token:");
  try {
    const response = await fetch(`${BASE_URL}/categories`);
    console.log(`   Status: ${response.status}`);
    if (!response.ok) {
      const error = await response.text();
      console.log(`   Error: ${error}`);
    }
  } catch (error) {
    console.log(`   Error: ${error.message}`);
  }

  // Test 2: Login to get a real JWT token
  console.log("\n2. Testing login to get JWT token:");
  try {
    const response = await fetch(`${BASE_URL}/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        email: "user@example.com",
        password: "test_password",
      }),
    });
    console.log(`   Status: ${response.status}`);
    if (response.ok) {
      const data = await response.json();
      console.log(`   Success: Login successful`);
      console.log(`   Token: ${data.token.substring(0, 50)}...`);

      // Test 3: Use the real JWT token
      console.log("\n3. Testing categories endpoint with real JWT token:");
      const categoriesResponse = await fetch(`${BASE_URL}/categories`, {
        headers: {
          Authorization: `Bearer ${data.token}`,
        },
      });
      console.log(`   Status: ${categoriesResponse.status}`);
      if (categoriesResponse.ok) {
        const categoriesData = await categoriesResponse.json();
        console.log(`   Success: Found ${categoriesData.length} categories`);
      } else {
        const error = await categoriesResponse.text();
        console.log(`   Error: ${error}`);
      }
    } else {
      const error = await response.text();
      console.log(`   Error: ${error}`);
    }
  } catch (error) {
    console.log(`   Error: ${error.message}`);
  }

  // Test 4: Test with invalid token
  console.log("\n4. Testing with invalid token:");
  try {
    const response = await fetch(`${BASE_URL}/categories`, {
      headers: {
        Authorization: "Bearer invalid-token-here",
      },
    });
    console.log(`   Status: ${response.status}`);
    if (!response.ok) {
      const error = await response.text();
      console.log(`   Error: ${error}`);
    }
  } catch (error) {
    console.log(`   Error: ${error.message}`);
  }
}

testJWTAuth();
