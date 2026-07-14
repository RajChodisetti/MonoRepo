"""Shared API request budget for ingestion providers."""

from __future__ import annotations


class BudgetExhausted(Exception):
    """Raised when an operation needs budget but none remains."""


class RequestBudget:
    def __init__(self, max_requests: int) -> None:
        if max_requests < 1:
            raise ValueError("max_requests must be at least 1")
        self.max_requests = max_requests
        self.total_used = 0

    @property
    def remaining(self) -> int:
        return max(0, self.max_requests - self.total_used)

    @property
    def exhausted(self) -> bool:
        return self.total_used >= self.max_requests

    def consume(self, count: int = 1) -> None:
        if count < 1:
            return
        if self.total_used + count > self.max_requests:
            raise BudgetExhausted(
                f"request budget exhausted ({self.total_used}/{self.max_requests} used, need {count})"
            )
        self.total_used += count

    def can_consume(self, count: int = 1) -> bool:
        return self.total_used + count <= self.max_requests
