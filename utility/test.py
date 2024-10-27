import aiohttp
import asyncio
import random
import string
import time

# Function to generate a random string of fixed length
def random_string(length=10):
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for _ in range(length))

# Function to generate a random email
def random_email():
    return f"{random_string(7)}@gmail.com"

# Function to send a single request
async def send_request(session):
    name = random_string(5)  # Generate a random name
    email = random_email()    # Generate a random email
    data = {"email": email, "name": name}
    
    async with session.post("http://localhost:4000/user_insert", json=data) as response:
        status = response.status
        if status == 200:
            print(f"Sent: {data}, Response: {status}")
        else:
            print(f"Error: {status} for {data}")

# Main function to control the load test
async def load_test(rate_per_second):
    async with aiohttp.ClientSession() as session:
        while True:  # Run indefinitely
            tasks = [send_request(session) for _ in range(rate_per_second)]
            await asyncio.gather(*tasks)
            # await asyncio.sleep(1)  # Sleep for 1 second before sending the next batch

if __name__ == "__main__":
    try:
        start_time = time.time()
        loop = asyncio.get_event_loop()
        loop.run_until_complete(load_test(10000))  # Adjust the rate as needed
    except KeyboardInterrupt:
        print("Load test stopped by keyboard.")
    finally:
        print(f"Total time running: {time.time() - start_time} seconds")
