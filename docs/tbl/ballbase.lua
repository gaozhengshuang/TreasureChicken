// Generated by github.com/davyxu/tabtoy
// Version: 2.8.10

module table {
export var TBall : table.ITBallDefine[] = [
		{ Id : 1, Atk : 1, Price : 1 	},
		{ Id : 2, Atk : 1, Price : 1 	}
	]


// Id
export var TBallById : game.Dictionary<table.ITBallDefine> = {}
function readTBallById(){
  for(let rec of TBall) {
    TBallById[rec.Id] = rec; 
  }
}
readTBallById();
}

