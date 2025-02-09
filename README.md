

### **GeeCache-s**

GeeCache-s is a distributed caching system inspired by [GeeCache](https://github.com/geektutu/7days-golang?tab=readme-ov-file#distributed-cache---geecache).  
This project serves as an educational and simplified implementation of a groupcache-like distributed cache, aiming to explore core concepts in distributed systems and caching mechanisms.

---

### **Implementation Details**

- **Communication Layer:**  
  Utilizes **HTTP** and **Protobuf** for lightweight and efficient communication between nodes as well as between clients and servers. This design ensures easy integration and supports serialization for structured data exchange.

- **Data Sharding with Consistent Hashing:**  
  The system implements consistent hashing to distribute keys across nodes. However, dynamic handling of node additions or removals is not yet supported and is planned for future development.

- **Caching Policies:**  
  The current implementation supports the **Least Recently Used (LRU)** and **Least Frequently Used (LFU)** caching policy. The design is modular, allowing for seamless extension to incorporate additional replacement strategies in the future.

---

### **Planned Improvements (TODO)**

- Add dynamic node handling for consistent hashing to better manage node additions and removals with minimal data movement.


