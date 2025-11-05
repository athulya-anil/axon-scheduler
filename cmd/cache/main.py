from fastapi import FastAPI

app = FastAPI()

@app.get("/")
def root():
    return {"service": "Axon Semantic Cache", "status": "ok"}
