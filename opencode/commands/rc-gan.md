---
description: rc — adversarial generator↔evaluator loop for subjective quality (Qwen3.7 Max, high)
agent: rc-gan
---

Drive the subjective quality of this target up to the target score: $ARGUMENTS

Use the `rc-gan` skill: write a weighted rubric, exercise the running artifact, score it with a fresh independent evaluator each round, and iterate on concrete defects until the threshold or a plateau. Keep the verification gate green every iteration. Correctness/security go through the review agent, not here.
