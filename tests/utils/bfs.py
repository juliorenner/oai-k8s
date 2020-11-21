def get_topology_graph():
    return {
        "node16": ["node14", "node15"],
        "node14": ["node3", "node4", "node5", "node16"],
        "node15": ["node3", "node4", "node5", "node16"],
        "node3": ["node6", "node7", "node8", "node14", "node15"],
        "node4": ["node6", "node10", "node11", "node14", "node15"],
        "node5": ["node7", "node11", "node14", "node15"],
        "node6": ["node3", "node4"],
        "node7": ["node3", "node5"],
        "node8": ["node3", "node9"],
        "node9": ["node8", "node10"],
        "node10": ["node9", "node4"],
        "node11": ["node5", "node4", "node12"],
        "node12": ["node11", "node13"],
        "node13": ["node12"]
    }

def find_shortest_path(start, dest):
    graph = get_topology_graph()
    queue = [[start]]
    visited = set()
    while len(queue) > 0:
        path = queue.pop(0)

        v = path[-1]

        if v == dest:
            return path
        
        if v not in visited:
            for n in graph[v]:
                new_path = list(path)
                new_path.append(n)
                queue.append(new_path)
            
            visited.add(v)

def count_hops(placement):
    hops_count = {}
    
    root = "node16"
    for v in placement:
        core_cu = find_shortest_path(root, placement[v]["cu"])
        cu_du = find_shortest_path(placement[v]["cu"], placement[v]["du"])
        du_ru = find_shortest_path(placement[v]["du"], placement[v]["ru"])

        hops_count[v] = len(core_cu)-1 + len(cu_du)-1 + len(du_ru)-1

    return hops_count

# a = {'1': {'cu': 'node9', 'du': 'node14', 'ru': 'node6', 'status': 'Running'}, '2': {'cu': 'node12', 'du': 'node8', 'ru': 'node7', 'status': 'Running'}, '3': {'cu': 'node11', 'du': 'node3', 'ru': 'node8', 'status': 'Running'}, '4': {'cu': 'node16', 'du': 'node15', 'ru': 'node10', 'status': 'Running'}, '5': {'cu': 'node4', 'du': 'node5', 'ru': 'node9', 'status': 'Running'}, '6': {'cu': 'node12', 'du': 'node12', 'ru': 'node11', 'status': 'Running'}, '7': {'cu': 'node16', 'du': 'node14', 'ru': 'node12', 'status': 'Running'}, '8': {'cu': 'node3', 'du': 'node15', 'ru': 'node13', 'status': 'Error'}}

# print(count_hops(a))

