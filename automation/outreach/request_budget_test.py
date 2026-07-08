"""Tests for request_budget."""

import unittest

from request_budget import BudgetExhausted, RequestBudget


class RequestBudgetTest(unittest.TestCase):
    def test_stops_at_limit(self):
        budget = RequestBudget(500)
        for _ in range(500):
            budget.consume(1)
        self.assertTrue(budget.exhausted)
        with self.assertRaises(BudgetExhausted):
            budget.consume(1)

    def test_can_consume(self):
        budget = RequestBudget(5)
        self.assertTrue(budget.can_consume(2))
        budget.consume(4)
        self.assertFalse(budget.can_consume(2))
        self.assertTrue(budget.can_consume(1))


if __name__ == "__main__":
    unittest.main()
