# Distributed Key-Value Store Project TODO List

## 1. Setup Basic Key-Value Store

### Issue 1: Implement In-Memory Key-Value Store
- [x] Create an in-memory key-value store.
- [x] Implement `Put`, `Get`, and `Delete` operations.
- [x] Use a map to store key-value pairs.

### Issue 2: Add Unit Tests for Basic Operations
- [ ] Write unit tests for `Put` operation.
- [ ] Write unit tests for `Get` operation.
- [ ] Write unit tests for `Delete` operation.
- [ ] Test edge cases like missing keys, overwriting values, etc.

---

## 2. Implement Concurrency

### Issue 3: Introduce Concurrency with Goroutines
- [x] Modify key-value store to support concurrent access.
- [x] Use goroutines to handle multiple operations simultaneously.

### Issue 4: Add Mutex for Concurrency Control
- [x] Protect shared data with `sync.Mutex` to avoid race conditions.
- [x] Ensure correctness with simultaneous read/write operations.

### Issue 5: Implement Read-Write Lock for Efficiency
- [x] Replace `sync.Mutex` with `sync.RWMutex`.
- [x] Allow multiple reads simultaneously, but limit write access.

---

## 3. Implement Basic Partitioning

### Issue 6: Create Multiple Partitions (Nodes)
- [x] Simulate multiple nodes by dividing the data into partitions.
- [x] Hash keys to determine which partition (node) they belong to.

### Issue 7: Implement Key Distribution Logic
- [x] Use a hash-based approach to distribute keys evenly across nodes.
- [ ] Ensure that keys are distributed fairly among partitions.

### Issue 8: Add Unit Tests for Partitioning
- [ ] Write unit tests to verify that keys are correctly partitioned.
- [ ] Test even distribution of keys across multiple nodes.

---

## 4. Add Replication

### Issue 9: Implement Basic Replication Mechanism
- [ ] Add asynchronous replication of key-value pairs to secondary nodes.
- [ ] Ensure that primary nodes replicate data to secondary nodes after writes.

### Issue 10: Ensure Consistency Across Replicas
- [ ] Implement consistency mechanisms between primary and replica nodes.
- [ ] Ensure data consistency without loss or duplication.

### Issue 11: Add Unit Tests for Replication
- [ ] Write unit tests to verify that replication works correctly.
- [ ] Ensure data is accurately replicated to secondary nodes.

---

## 5. Handle Node Failures

### Issue 12: Simulate Node Failures
- [ ] Implement logic to simulate node failures.
- [ ] Test how the system handles simulated node failures.

### Issue 13: Implement Failure Handling and Recovery
- [ ] Add logic for handling node failures.
- [ ] Implement retry mechanisms or request redirection in case of failures.
- [ ] Ensure that the system can recover without significant disruption.

### Issue 14: Add Unit Tests for Failure Handling
- [ ] Write unit tests to verify that node failures are handled properly.
- [ ] Ensure the system can recover from failures as expected.

---

## 6. Concurrency Testing and Performance Optimization

### Issue 15: Perform Concurrency Testing
- [ ] Test the key-value store under heavy concurrent load.
- [ ] Identify performance bottlenecks and race conditions.

### Issue 16: Optimize Performance
- [ ] Improve efficiency of concurrency controls.
- [ ] Optimize replication logic for better performance.
- [ ] Implement any other necessary performance optimizations.

### Issue 17: Write Documentation
- [ ] Document the design of the key-value store.
- [ ] Explain the implementation details and how to use the store.
- [ ] Provide guidance on extending and contributing to the project.